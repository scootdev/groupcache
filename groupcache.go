/*
Copyright 2012 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package groupcache provides a data loading mechanism with caching
// and de-duplication that works across a set of peer processes.
//
// Each data Get first consults its local cache, otherwise delegates
// to the requested key's canonical owner, which then checks its cache
// or finally gets the data.  In the common case, many concurrent
// cache misses across a set of peers for the same key result in just
// one cache fill.
//
// Put supports manual loading of a data's canonical owner's cache.
package groupcache

import (
	"errors"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/ptypes"
	pb "github.com/twitter/groupcache/groupcachepb"
	"github.com/twitter/groupcache/lru"
	"github.com/twitter/groupcache/singleflight"
)

// A Getter loads data for a key.
type Getter interface {
	// Get returns the value identified by key, populating dest.
	//
	// The returned data must be unversioned. That is, key must
	// uniquely describe the loaded data, without an implicit
	// current time, and without relying on cache expiration
	// mechanisms.
	Get(ctx Context, key string, dest Sink) (*time.Time, error)
}

// A GetterFunc implements Getter with a function.
type GetterFunc func(ctx Context, key string, dest Sink) (*time.Time, error)

func (f GetterFunc) Get(ctx Context, key string, dest Sink) (*time.Time, error) {
	return f(ctx, key, dest)
}

// A Putter stores data for a key.
type Putter interface {
	// Put stores the data identified by key in the cache.
	//
	// Data to be stored must be unversioned as per Getters.
	// Data cannot be invalidated - it is assumed that
	// putting any data with a preexisting key can be
	// interpreted as a no-op.
	// TTL is a duration value that will be
	// passed through to any underlying PutterFunc.
	Put(ctx Context, key string, data []byte, ttl time.Duration) error
}

// A PutterFunc implements Putter with a function.
type PutterFunc func(ctx Context, key string, data []byte, ttl time.Duration) error

func (f PutterFunc) Put(ctx Context, key string, data []byte, ttl time.Duration) error {
	return f(ctx, key, data, ttl)
}

// A GetterPutter combines the Getter and Putter interfaces.
type GetterPutter interface {
	Getter
	Putter
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)

	initPeerServerOnce sync.Once
	initPeerServer     func()
)

// GetGroup returns the named group previously created with NewGroup, or
// nil if there's no such group.
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// NewGroup creates a coordinated group-aware Getter/Putter.
//
// The returned Getter/Putter tries (but does not guarantee) to run only one
// Get/Put call at once for a given key across an entire set of peer
// processes. Concurrent callers both in the local process and in
// other processes receive copies of the answer once the original Get/Put
// completes.
//
// The group name must be unique for each getter/putter.
func NewGroup(name string, cacheBytes int64, getter Getter, putter Putter) *Group {
	return newGroup(name, cacheBytes, getter, putter, nil)
}

// If peers is nil, the peerPicker is called via a sync.Once to initialize it.
func newGroup(name string, cacheBytes int64, getter Getter, putter Putter, peers PeerPicker) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	if putter == nil {
		panic("nil Putter")
	}
	mu.Lock()
	defer mu.Unlock()
	initPeerServerOnce.Do(callInitPeerServer)
	if _, dup := groups[name]; dup {
		panic("duplicate registration of group " + name)
	}
	g := &Group{
		name:       name,
		getter:     getter,
		putter:     putter,
		peers:      peers,
		cacheBytes: cacheBytes,
		loadGroup:  &singleflight.Group{},
	}
	if fn := newGroupHook; fn != nil {
		fn(g)
	}
	groups[name] = g
	return g
}

// newGroupHook, if non-nil, is called right after a new group is created.
var newGroupHook func(*Group)

// RegisterNewGroupHook registers a hook that is run each time
// a group is created.
func RegisterNewGroupHook(fn func(*Group)) {
	if newGroupHook != nil {
		panic("RegisterNewGroupHook called more than once")
	}
	newGroupHook = fn
}

// RegisterServerStart registers a hook that is run when the first
// group is created.
func RegisterServerStart(fn func()) {
	if initPeerServer != nil {
		panic("RegisterServerStart called more than once")
	}
	initPeerServer = fn
}

func callInitPeerServer() {
	if initPeerServer != nil {
		initPeerServer()
	}
}

// A Group is a cache namespace and associated data loaded spread over
// a group of 1 or more machines.
type Group struct {
	name       string
	getter     Getter
	putter     Putter
	peersOnce  sync.Once
	peers      PeerPicker
	cacheBytes int64 // limit for sum of mainCache and hotCache size

	// mainCache is a cache of the keys for which this process
	// (amongst its peers) is authoritative. That is, this cache
	// contains keys which consistent hash on to this process's
	// peer number.
	mainCache cache

	// hotCache contains keys/values for which this peer is not
	// authoritative (otherwise they would be in mainCache), but
	// are popular enough to warrant mirroring in this process to
	// avoid going over the network to fetch from a peer.  Having
	// a hotCache avoids network hotspotting, where a peer's
	// network card could become the bottleneck on a popular key.
	// This cache is used sparingly to maximize the total number
	// of key/value pairs that can be stored globally.
	hotCache cache

	// loadGroup ensures that each key is only fetched once
	// (either locally or remotely), regardless of the number of
	// concurrent callers.
	loadGroup flightGroup

	_ int32 // force Stats to be 8-byte aligned on 32-bit platforms

	// Stats are statistics on the group.
	Stats Stats
}

// flightGroup is defined as an interface which flightgroup.Group
// satisfies.  We define this so that we may test with an alternate
// implementation.
type flightGroup interface {
	// Done is called when Do is done.
	Do(key string, fn func() (interface{}, error)) (interface{}, error)
}

// Stats are per-group statistics.
type Stats struct {
	Gets           AtomicInt // any Get request, including from peers
	Puts           AtomicInt // any Put request, including from peers
	CacheHits      AtomicInt // either cache was good
	Loads          AtomicInt // Gets not from the cache
	LoadsDeduped   AtomicInt // after singleflight
	LocalLoads     AtomicInt // total good local loads
	LocalLoadErrs  AtomicInt // total bad local loads
	Stores         AtomicInt // Puts that weren't in the cache
	StoresDeduped  AtomicInt // after singleflight
	LocalStores    AtomicInt // total good local stores
	LocalStoreErrs AtomicInt // total bad local stores
	PeerStores     AtomicInt // either remote store or remote cache hit (not an error)
	PeerLoads      AtomicInt // either remote load or remote cache hit (not an error)
	PeerErrors     AtomicInt
	ServerRequests AtomicInt // requests that came over the network from peers
}

// Name returns the name of the group.
func (g *Group) Name() string {
	return g.name
}

func (g *Group) initPeers() {
	if g.peers == nil {
		g.peers = getPeers(g.name)
	}
}

// Get functions

func (g *Group) Get(ctx Context, key string, dest Sink) (*time.Time, error) {
	g.peersOnce.Do(g.initPeers)
	g.Stats.Gets.Add(1)
	if dest == nil {
		return nil, errors.New("groupcache: nil dest Sink")
	}
	value, expiration, cacheHit := g.lookupCache(key)

	if cacheHit {
		g.Stats.CacheHits.Add(1)
		return expiration, setSinkView(dest, value)
	}

	// Optimization to avoid double unmarshalling or copying: keep
	// track of whether the dest was already populated. One caller
	// (if local) will set this; the losers will not. The common
	// case will likely be one caller.
	destPopulated := false
	value, destPopulated, expiration, err := g.load(ctx, key, dest)
	if err != nil {
		return nil, err
	}
	if destPopulated {
		return expiration, nil
	}
	return expiration, setSinkView(dest, value)
}

type loadResult struct {
	value      ByteView
	expiration *time.Time
}

// underlying Get logic - loads key either by invoking the getter locally or by sending it to another machine.
func (g *Group) load(ctx Context, key string, dest Sink) (value ByteView, destPopulated bool, expiration *time.Time, err error) {
	g.Stats.Loads.Add(1)
	viewi, err := g.loadGroup.Do(key, func() (interface{}, error) {
		// Check the cache again because singleflight can only dedup calls
		// that overlap concurrently.  It's possible for 2 concurrent
		// requests to miss the cache, resulting in 2 load() calls.  An
		// unfortunate goroutine scheduling would result in this callback
		// being run twice, serially.  If we don't check the cache again,
		// cache.nbytes would be incremented below even though there will
		// be only one entry for this key.
		//
		// Consider the following serialized event ordering for two
		// goroutines in which this callback gets called twice for the
		// same key:
		// 1: Get("key")
		// 2: Get("key")
		// 1: lookupCache("key")
		// 2: lookupCache("key")
		// 1: load("key")
		// 2: load("key")
		// 1: loadGroup.Do("key", fn)
		// 1: fn()
		// 2: loadGroup.Do("key", fn)
		// 2: fn()
		if value, expiration, cacheHit := g.lookupCache(key); cacheHit {
			g.Stats.CacheHits.Add(1)
			return loadResult{value: value, expiration: expiration}, nil
		}
		g.Stats.LoadsDeduped.Add(1)
		var value ByteView
		var expiration *time.Time
		var err error
		if peer, ok := g.peers.PickPeer(key); ok {
			value, expiration, err = g.getFromPeer(ctx, peer, key)
			if err == nil {
				g.Stats.PeerLoads.Add(1)
				return loadResult{value: value, expiration: expiration}, nil
			}
			g.Stats.PeerErrors.Add(1)
			// TODO(bradfitz): log the peer's error? keep
			// log of the past few for /groupcachez?  It's
			// probably boring (normal task movement), so not
			// worth logging I imagine.
		}
		value, expiration, err = g.getLocally(ctx, key, dest)
		if err != nil {
			g.Stats.LocalLoadErrs.Add(1)
			return nil, err
		}
		g.Stats.LocalLoads.Add(1)
		destPopulated = true // only one caller of load gets this return value
		g.populateCache(key, value, &g.mainCache, expiration)
		return loadResult{value: value, expiration: expiration}, nil
	})
	if err == nil {
		lr, ok := viewi.(loadResult)
		if !ok {
			err = errors.New("groupcache: failed interface conversion")
		} else {
			value = lr.value
			expiration = lr.expiration
		}
	}
	return
}

func (g *Group) getLocally(ctx Context, key string, dest Sink) (ByteView, *time.Time, error) {
	expiration, err := g.getter.Get(ctx, key, dest)
	if err != nil {
		return ByteView{}, nil, err
	}
	dv, err := dest.view()
	return dv, expiration, err
}

func (g *Group) getFromPeer(ctx Context, peer ProtoPeer, key string) (ByteView, *time.Time, error) {
	req := &pb.GetRequest{
		Group: g.name,
		Key:   key,
	}
	res := &pb.GetResponse{}
	err := peer.Get(ctx, req, res)
	if err != nil {
		return ByteView{}, nil, err
	}
	value := ByteView{b: res.Value}
	expiration, err := ptypes.Timestamp(res.Expiration)
	if err != nil {
		return ByteView{}, nil, err
	}
	expiration = expiration.UTC()
	// TODO(bradfitz): use res.MinuteQps or something smart to
	// conditionally populate hotCache.  For now just do it some
	// percentage of the time.
	if rand.Intn(10) == 0 {
		g.populateCache(key, value, &g.hotCache, &expiration)
	}
	return value, &expiration, nil
}

// Put functions

func (g *Group) Put(ctx Context, key string, data []byte, ttl time.Duration) error {
	g.peersOnce.Do(g.initPeers)
	g.Stats.Puts.Add(1)
	if data == nil {
		return errors.New("groupcache: nil data")
	}
	_, _, cacheHit := g.lookupCache(key)

	if cacheHit {
		g.Stats.CacheHits.Add(1)
		return nil
	}

	err := g.store(ctx, key, data, ttl)
	if err != nil {
		return err
	}
	return nil
}

// underlying Put logic - stores data for key either by invoking the putter locally or by sending it to another machine.
func (g *Group) store(ctx Context, key string, data []byte, ttl time.Duration) (err error) {
	g.Stats.Stores.Add(1)
	_, err = g.loadGroup.Do(key, func() (interface{}, error) {
		// Deduplication checks - see explanation in load()
		if _, _, cacheHit := g.lookupCache(key); cacheHit {
			g.Stats.CacheHits.Add(1)
			return nil, nil
		}
		g.Stats.StoresDeduped.Add(1)
		var err error
		if peer, ok := g.peers.PickPeer(key); ok {
			err = g.putFromPeer(ctx, peer, key, data, ttl)
			if err == nil {
				g.Stats.PeerStores.Add(1)
				return nil, nil
			}
			g.Stats.PeerErrors.Add(1)
		}
		err = g.putLocally(ctx, key, data, ttl)
		if err != nil {
			g.Stats.LocalStoreErrs.Add(1)
			return nil, err
		}
		g.Stats.LocalStores.Add(1)
		value := ByteView{b: data}
		expiration := time.Now().UTC().Add(ttl)
		g.populateCache(key, value, &g.mainCache, &expiration)
		return nil, nil
	})
	return
}

func (g *Group) putLocally(ctx Context, key string, data []byte, ttl time.Duration) error {
	return g.putter.Put(ctx, key, data, ttl)
}

func (g *Group) putFromPeer(ctx Context, peer ProtoPeer, key string, data []byte, ttl time.Duration) error {
	req := &pb.PutRequest{
		Group: g.name,
		Key:   key,
		Value: data,
		Ttl:   ptypes.DurationProto(ttl),
	}
	res := &pb.PutResponse{}
	err := peer.Put(ctx, req, res)
	if err != nil {
		return err
	}
	value := ByteView{b: data}
	// TODO(bradfitz): use res.MinuteQps or something smart to
	// conditionally populate hotCache.  For now just do it some
	// percentage of the time.
	if rand.Intn(10) == 0 {
		expiration := time.Now().UTC().Add(ttl)
		g.populateCache(key, value, &g.hotCache, &expiration)
	}
	return nil
}

// Cache utils

func (g *Group) lookupCache(key string) (value ByteView, expiration *time.Time, ok bool) {
	if g.cacheBytes <= 0 {
		return
	}
	value, expiration, ok = g.mainCache.get(key)
	if ok {
		return
	}
	value, expiration, ok = g.hotCache.get(key)
	return
}

func (g *Group) populateCache(key string, value ByteView, cache *cache, expiration *time.Time) {
	if g.cacheBytes <= 0 {
		return
	}
	cache.add(key, value, expiration)

	// Evict items from cache(s) if necessary.
	for {
		mainBytes := g.mainCache.bytes()
		hotBytes := g.hotCache.bytes()
		if mainBytes+hotBytes <= g.cacheBytes {
			return
		}

		// TODO(bradfitz): this is good-enough-for-now logic.
		// It should be something based on measurements and/or
		// respecting the costs of different resources.
		victim := &g.mainCache
		if hotBytes > mainBytes/8 {
			victim = &g.hotCache
		}
		victim.removeOldest()
	}
}

// CacheType represents a type of cache.
type CacheType int

const (
	// The MainCache is the cache for items that this peer is the
	// owner for.
	MainCache CacheType = iota + 1

	// The HotCache is the cache for items that seem popular
	// enough to replicate to this node, even though it's not the
	// owner.
	HotCache
)

// CacheStats returns stats about the provided cache within the group.
func (g *Group) CacheStats(which CacheType) CacheStats {
	switch which {
	case MainCache:
		return g.mainCache.stats()
	case HotCache:
		return g.hotCache.stats()
	default:
		return CacheStats{}
	}
}

// cache is a wrapper around an *lru.Cache that adds synchronization,
// makes values always be ByteView, and counts the size of all keys and
// values.
type cache struct {
	mu         sync.RWMutex
	nbytes     int64 // of all keys and values
	lru        *lru.Cache
	nhit, nget int64
	nevict     int64 // number of evictions
	metadata   map[string]*cacheValueMetadata
}

type cacheValueMetadata struct {
	expiration *time.Time
}

func (c *cache) getMetadata(key string) *cacheValueMetadata {
	if c.metadata == nil {
		c.metadata = make(map[string]*cacheValueMetadata)
	}
	m := c.metadata[key]
	if m == nil {
		m = &cacheValueMetadata{}
		c.metadata[key] = m
	}
	return m
}

func (c *cacheValueMetadata) addExpiration(t *time.Time) {
	if t != nil {
		utc := (*t).UTC()
		c.expiration = &utc
	}
}

func (c *cache) stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return CacheStats{
		Bytes:     c.nbytes,
		Items:     c.itemsLocked(),
		Gets:      c.nget,
		Hits:      c.nhit,
		Evictions: c.nevict,
	}
}

func (c *cache) add(key string, value ByteView, expiration *time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.getMetadata(key).addExpiration(expiration)
	if c.lru == nil {
		c.lru = &lru.Cache{
			OnEvicted: func(key lru.Key, value interface{}) {
				val := value.(ByteView)
				delete(c.metadata, key.(string))
				c.nbytes -= int64(len(key.(string))) + int64(val.Len())
				c.nevict++
			},
		}
	}
	c.lru.Add(key, value)
	c.nbytes += int64(len(key)) + int64(value.Len())
}

func (c *cache) get(key string) (value ByteView, expiration *time.Time, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.nget++
	if c.lru == nil {
		return
	}
	if expiration = c.getMetadata(key).expiration; expiration != nil && expiration.Before(time.Now().UTC()) {
		return
	}
	vi, ok := c.lru.Get(key)
	if !ok {
		return
	}
	c.nhit++
	return vi.(ByteView), expiration, true
}

func (c *cache) removeOldest() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru != nil {
		c.lru.RemoveOldest()
	}
}

func (c *cache) bytes() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.nbytes
}

func (c *cache) items() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.itemsLocked()
}

func (c *cache) itemsLocked() int64 {
	if c.lru == nil {
		return 0
	}
	return int64(c.lru.Len())
}

// An AtomicInt is an int64 to be accessed atomically.
type AtomicInt int64

// Add atomically adds n to i.
func (i *AtomicInt) Add(n int64) {
	atomic.AddInt64((*int64)(i), n)
}

// Get atomically gets the value of i.
func (i *AtomicInt) Get() int64 {
	return atomic.LoadInt64((*int64)(i))
}

func (i *AtomicInt) String() string {
	return strconv.FormatInt(i.Get(), 10)
}

// CacheStats are returned by stats accessors on Group.
type CacheStats struct {
	Bytes     int64
	Items     int64
	Gets      int64
	Hits      int64
	Evictions int64
}
