// Code generated by protoc-gen-go. DO NOT EDIT.
// source: groupcache.proto

package groupcachepb

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type GetRequest struct {
	Group                string   `protobuf:"bytes,1,opt,name=group,proto3" json:"group,omitempty"`
	Key                  string   `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GetRequest) Reset()         { *m = GetRequest{} }
func (m *GetRequest) String() string { return proto.CompactTextString(m) }
func (*GetRequest) ProtoMessage()    {}
func (*GetRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_aba3e029bfa16198, []int{0}
}

func (m *GetRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetRequest.Unmarshal(m, b)
}
func (m *GetRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetRequest.Marshal(b, m, deterministic)
}
func (m *GetRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetRequest.Merge(m, src)
}
func (m *GetRequest) XXX_Size() int {
	return xxx_messageInfo_GetRequest.Size(m)
}
func (m *GetRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_GetRequest.DiscardUnknown(m)
}

var xxx_messageInfo_GetRequest proto.InternalMessageInfo

func (m *GetRequest) GetGroup() string {
	if m != nil {
		return m.Group
	}
	return ""
}

func (m *GetRequest) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

type GetResponse struct {
	Value                []byte               `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
	MinuteQps            float64              `protobuf:"fixed64,2,opt,name=minute_qps,json=minuteQps,proto3" json:"minute_qps,omitempty"`
	Ttl                  *timestamp.Timestamp `protobuf:"bytes,3,opt,name=ttl,proto3" json:"ttl,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *GetResponse) Reset()         { *m = GetResponse{} }
func (m *GetResponse) String() string { return proto.CompactTextString(m) }
func (*GetResponse) ProtoMessage()    {}
func (*GetResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_aba3e029bfa16198, []int{1}
}

func (m *GetResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GetResponse.Unmarshal(m, b)
}
func (m *GetResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GetResponse.Marshal(b, m, deterministic)
}
func (m *GetResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GetResponse.Merge(m, src)
}
func (m *GetResponse) XXX_Size() int {
	return xxx_messageInfo_GetResponse.Size(m)
}
func (m *GetResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_GetResponse.DiscardUnknown(m)
}

var xxx_messageInfo_GetResponse proto.InternalMessageInfo

func (m *GetResponse) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *GetResponse) GetMinuteQps() float64 {
	if m != nil {
		return m.MinuteQps
	}
	return 0
}

func (m *GetResponse) GetTtl() *timestamp.Timestamp {
	if m != nil {
		return m.Ttl
	}
	return nil
}

type PutRequest struct {
	Group                string               `protobuf:"bytes,1,opt,name=group,proto3" json:"group,omitempty"`
	Key                  string               `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty"`
	Value                []byte               `protobuf:"bytes,3,opt,name=value,proto3" json:"value,omitempty"`
	Ttl                  *timestamp.Timestamp `protobuf:"bytes,4,opt,name=ttl,proto3" json:"ttl,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *PutRequest) Reset()         { *m = PutRequest{} }
func (m *PutRequest) String() string { return proto.CompactTextString(m) }
func (*PutRequest) ProtoMessage()    {}
func (*PutRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_aba3e029bfa16198, []int{2}
}

func (m *PutRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PutRequest.Unmarshal(m, b)
}
func (m *PutRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PutRequest.Marshal(b, m, deterministic)
}
func (m *PutRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PutRequest.Merge(m, src)
}
func (m *PutRequest) XXX_Size() int {
	return xxx_messageInfo_PutRequest.Size(m)
}
func (m *PutRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_PutRequest.DiscardUnknown(m)
}

var xxx_messageInfo_PutRequest proto.InternalMessageInfo

func (m *PutRequest) GetGroup() string {
	if m != nil {
		return m.Group
	}
	return ""
}

func (m *PutRequest) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func (m *PutRequest) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *PutRequest) GetTtl() *timestamp.Timestamp {
	if m != nil {
		return m.Ttl
	}
	return nil
}

type PutResponse struct {
	MinuteQps            float64  `protobuf:"fixed64,1,opt,name=minute_qps,json=minuteQps,proto3" json:"minute_qps,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PutResponse) Reset()         { *m = PutResponse{} }
func (m *PutResponse) String() string { return proto.CompactTextString(m) }
func (*PutResponse) ProtoMessage()    {}
func (*PutResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_aba3e029bfa16198, []int{3}
}

func (m *PutResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PutResponse.Unmarshal(m, b)
}
func (m *PutResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PutResponse.Marshal(b, m, deterministic)
}
func (m *PutResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PutResponse.Merge(m, src)
}
func (m *PutResponse) XXX_Size() int {
	return xxx_messageInfo_PutResponse.Size(m)
}
func (m *PutResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_PutResponse.DiscardUnknown(m)
}

var xxx_messageInfo_PutResponse proto.InternalMessageInfo

func (m *PutResponse) GetMinuteQps() float64 {
	if m != nil {
		return m.MinuteQps
	}
	return 0
}

func init() {
	proto.RegisterType((*GetRequest)(nil), "groupcachepb.GetRequest")
	proto.RegisterType((*GetResponse)(nil), "groupcachepb.GetResponse")
	proto.RegisterType((*PutRequest)(nil), "groupcachepb.PutRequest")
	proto.RegisterType((*PutResponse)(nil), "groupcachepb.PutResponse")
}

func init() { proto.RegisterFile("groupcache.proto", fileDescriptor_aba3e029bfa16198) }

var fileDescriptor_aba3e029bfa16198 = []byte{
	// 274 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x91, 0xb1, 0x4e, 0xc3, 0x30,
	0x10, 0x86, 0x6b, 0x02, 0x48, 0xbd, 0x74, 0xa8, 0x2c, 0x86, 0x10, 0x09, 0x51, 0x79, 0xea, 0x50,
	0xb9, 0x52, 0x61, 0x64, 0x63, 0xc8, 0x5a, 0x2c, 0x76, 0x94, 0x54, 0x47, 0xa8, 0x48, 0x6a, 0xb7,
	0x3e, 0x23, 0x78, 0x03, 0x1e, 0x1b, 0xd9, 0x6e, 0x95, 0x50, 0x18, 0x60, 0xb3, 0xcf, 0xfe, 0xf4,
	0x7f, 0xbf, 0x0d, 0xe3, 0x7a, 0xa7, 0x9d, 0x59, 0x95, 0xab, 0x17, 0x94, 0x66, 0xa7, 0x49, 0xf3,
	0x51, 0x37, 0x31, 0x55, 0x7e, 0x5d, 0x6b, 0x5d, 0x37, 0x38, 0x0f, 0x67, 0x95, 0x7b, 0x9e, 0xd3,
	0xba, 0x45, 0x4b, 0x65, 0x6b, 0xe2, 0x75, 0x71, 0x0b, 0x50, 0x20, 0x29, 0xdc, 0x3a, 0xb4, 0xc4,
	0x2f, 0xe0, 0x2c, 0xe0, 0x19, 0x9b, 0xb0, 0xe9, 0x50, 0xc5, 0x0d, 0x1f, 0x43, 0xf2, 0x8a, 0x1f,
	0xd9, 0x49, 0x98, 0xf9, 0xa5, 0x30, 0x90, 0x06, 0xca, 0x1a, 0xbd, 0xb1, 0xe8, 0xb1, 0xb7, 0xb2,
	0x71, 0x18, 0xb0, 0x91, 0x8a, 0x1b, 0x7e, 0x05, 0xd0, 0xae, 0x37, 0x8e, 0xf0, 0x69, 0x6b, 0x6c,
	0xa0, 0x99, 0x1a, 0xc6, 0xc9, 0x83, 0xb1, 0x7c, 0x06, 0x09, 0x51, 0x93, 0x25, 0x13, 0x36, 0x4d,
	0x17, 0xb9, 0x8c, 0xa2, 0xf2, 0x20, 0x2a, 0x1f, 0x0f, 0xa2, 0xca, 0x5f, 0x13, 0xef, 0x00, 0x4b,
	0xf7, 0x5f, 0xcf, 0x4e, 0x2c, 0xe9, 0x8b, 0xed, 0x93, 0x4f, 0xff, 0x96, 0x3c, 0x83, 0x34, 0x24,
	0xef, 0xbb, 0x7e, 0x6f, 0xc5, 0x8e, 0x5a, 0x2d, 0x3e, 0x19, 0x40, 0xe1, 0x6d, 0xee, 0xfd, 0x0f,
	0xf0, 0x3b, 0x48, 0x0a, 0x24, 0x9e, 0xc9, 0xfe, 0xaf, 0xc8, 0xee, 0xc5, 0xf3, 0xcb, 0x5f, 0x4e,
	0x62, 0x92, 0x18, 0x78, 0x7a, 0xe9, 0x7e, 0xd0, 0xdd, 0x3b, 0x1c, 0xd3, 0x3d, 0x4f, 0x31, 0xa8,
	0xce, 0x43, 0xa3, 0x9b, 0xaf, 0x00, 0x00, 0x00, 0xff, 0xff, 0x90, 0xc7, 0x70, 0x0e, 0x24, 0x02,
	0x00, 0x00,
}
