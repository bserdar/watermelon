// Code generated by protoc-gen-go. DO NOT EDIT.
// source: host.proto

package pb

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
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
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Address struct {
	// name associated with this address. primary, primary4, primary6, etc.
	Name string `protobuf:"bytes,1,opt,name=Name,proto3" json:"Name,omitempty"`
	// The address
	Address              string   `protobuf:"bytes,2,opt,name=Address,proto3" json:"Address,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Address) Reset()         { *m = Address{} }
func (m *Address) String() string { return proto.CompactTextString(m) }
func (*Address) ProtoMessage()    {}
func (*Address) Descriptor() ([]byte, []int) {
	return fileDescriptor_85e40b83b4d50a8d, []int{0}
}

func (m *Address) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Address.Unmarshal(m, b)
}
func (m *Address) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Address.Marshal(b, m, deterministic)
}
func (m *Address) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Address.Merge(m, src)
}
func (m *Address) XXX_Size() int {
	return xxx_messageInfo_Address.Size(m)
}
func (m *Address) XXX_DiscardUnknown() {
	xxx_messageInfo_Address.DiscardUnknown(m)
}

var xxx_messageInfo_Address proto.InternalMessageInfo

func (m *Address) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Address) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

// HostInfo describes a host.
type HostInfo struct {
	// HostId is the unique identifier for this host. It could be fqdn,
	// ip, or any other symbolic name
	ID        string     `protobuf:"bytes,1,opt,name=ID,proto3" json:"ID,omitempty"`
	Addresses []*Address `protobuf:"bytes,3,rep,name=Addresses,proto3" json:"Addresses,omitempty"`
	// labels associated with this host
	Labels []string `protobuf:"bytes,4,rep,name=Labels,proto3" json:"Labels,omitempty"`
	// key-value pairs associated with this host
	Properties           map[string]string `protobuf:"bytes,5,rep,name=Properties,proto3" json:"Properties,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *HostInfo) Reset()         { *m = HostInfo{} }
func (m *HostInfo) String() string { return proto.CompactTextString(m) }
func (*HostInfo) ProtoMessage()    {}
func (*HostInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_85e40b83b4d50a8d, []int{1}
}

func (m *HostInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HostInfo.Unmarshal(m, b)
}
func (m *HostInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HostInfo.Marshal(b, m, deterministic)
}
func (m *HostInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HostInfo.Merge(m, src)
}
func (m *HostInfo) XXX_Size() int {
	return xxx_messageInfo_HostInfo.Size(m)
}
func (m *HostInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_HostInfo.DiscardUnknown(m)
}

var xxx_messageInfo_HostInfo proto.InternalMessageInfo

func (m *HostInfo) GetID() string {
	if m != nil {
		return m.ID
	}
	return ""
}

func (m *HostInfo) GetAddresses() []*Address {
	if m != nil {
		return m.Addresses
	}
	return nil
}

func (m *HostInfo) GetLabels() []string {
	if m != nil {
		return m.Labels
	}
	return nil
}

func (m *HostInfo) GetProperties() map[string]string {
	if m != nil {
		return m.Properties
	}
	return nil
}

// CommandError
type CommandError struct {
	Host                 string   `protobuf:"bytes,1,opt,name=Host,proto3" json:"Host,omitempty"`
	Msg                  string   `protobuf:"bytes,2,opt,name=Msg,proto3" json:"Msg,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *CommandError) Reset()         { *m = CommandError{} }
func (m *CommandError) String() string { return proto.CompactTextString(m) }
func (*CommandError) ProtoMessage()    {}
func (*CommandError) Descriptor() ([]byte, []int) {
	return fileDescriptor_85e40b83b4d50a8d, []int{2}
}

func (m *CommandError) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CommandError.Unmarshal(m, b)
}
func (m *CommandError) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CommandError.Marshal(b, m, deterministic)
}
func (m *CommandError) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CommandError.Merge(m, src)
}
func (m *CommandError) XXX_Size() int {
	return xxx_messageInfo_CommandError.Size(m)
}
func (m *CommandError) XXX_DiscardUnknown() {
	xxx_messageInfo_CommandError.DiscardUnknown(m)
}

var xxx_messageInfo_CommandError proto.InternalMessageInfo

func (m *CommandError) GetHost() string {
	if m != nil {
		return m.Host
	}
	return ""
}

func (m *CommandError) GetMsg() string {
	if m != nil {
		return m.Msg
	}
	return ""
}

func init() {
	proto.RegisterType((*Address)(nil), "pb.Address")
	proto.RegisterType((*HostInfo)(nil), "pb.HostInfo")
	proto.RegisterMapType((map[string]string)(nil), "pb.HostInfo.PropertiesEntry")
	proto.RegisterType((*CommandError)(nil), "pb.CommandError")
}

func init() { proto.RegisterFile("host.proto", fileDescriptor_85e40b83b4d50a8d) }

var fileDescriptor_85e40b83b4d50a8d = []byte{
	// 293 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x51, 0x41, 0x4b, 0xf3, 0x40,
	0x10, 0x25, 0x49, 0xdb, 0xef, 0xeb, 0x54, 0x54, 0x16, 0x91, 0x45, 0x3c, 0x94, 0x9e, 0xea, 0x25,
	0x91, 0x2a, 0x28, 0xa2, 0x07, 0xb5, 0x05, 0x0b, 0x2a, 0x92, 0xa3, 0xb7, 0xdd, 0x66, 0x6c, 0x8b,
	0x49, 0x76, 0x99, 0xdd, 0x46, 0xf2, 0x57, 0xfd, 0x35, 0xb2, 0x71, 0x63, 0xc5, 0xdb, 0x7b, 0xf3,
	0x66, 0xde, 0xec, 0xbe, 0x01, 0x58, 0x29, 0x63, 0x63, 0x4d, 0xca, 0x2a, 0x16, 0x6a, 0x39, 0xba,
	0x80, 0x7f, 0xb7, 0x59, 0x46, 0x68, 0x0c, 0x63, 0xd0, 0x79, 0x16, 0x05, 0xf2, 0x60, 0x18, 0x8c,
	0xfb, 0x69, 0x83, 0x19, 0xff, 0x91, 0x79, 0xd8, 0x94, 0x5b, 0x3a, 0xfa, 0x0c, 0xe0, 0xff, 0x83,
	0x32, 0x76, 0x5e, 0xbe, 0x29, 0xb6, 0x0b, 0xe1, 0x7c, 0xea, 0x07, 0xc3, 0xf9, 0x94, 0x9d, 0x40,
	0xdf, 0xf7, 0xa1, 0xe1, 0xd1, 0x30, 0x1a, 0x0f, 0x26, 0x83, 0x58, 0xcb, 0xd8, 0x17, 0xd3, 0xad,
	0xca, 0x0e, 0xa1, 0xf7, 0x28, 0x24, 0xe6, 0x86, 0x77, 0x86, 0xd1, 0xb8, 0x9f, 0x7a, 0xc6, 0xae,
	0x01, 0x5e, 0x48, 0x69, 0x24, 0xbb, 0x46, 0xc3, 0xbb, 0x8d, 0xc7, 0xb1, 0xf3, 0x68, 0x97, 0xc6,
	0x5b, 0x79, 0x56, 0x5a, 0xaa, 0xd3, 0x5f, 0xfd, 0x47, 0x37, 0xb0, 0xf7, 0x47, 0x66, 0xfb, 0x10,
	0xbd, 0x63, 0xed, 0x1f, 0xe9, 0x20, 0x3b, 0x80, 0x6e, 0x25, 0xf2, 0x0d, 0xfa, 0xaf, 0x7d, 0x93,
	0xab, 0xf0, 0x32, 0x18, 0x9d, 0xc3, 0xce, 0xbd, 0x2a, 0x0a, 0x51, 0x66, 0x33, 0x22, 0x45, 0x2e,
	0x1a, 0xb7, 0xb6, 0x8d, 0xc6, 0x61, 0xe7, 0xf7, 0x64, 0x96, 0x7e, 0xd6, 0xc1, 0xbb, 0xc9, 0xeb,
	0xe9, 0x72, 0x6d, 0x57, 0x1b, 0x19, 0x2f, 0x54, 0x91, 0x2c, 0x72, 0xb5, 0xc9, 0x34, 0xad, 0x2b,
	0xb1, 0xa8, 0x73, 0x21, 0x4d, 0xf2, 0x21, 0x2c, 0x52, 0x81, 0xb9, 0x2a, 0x13, 0x83, 0x54, 0x21,
	0x25, 0x5a, 0xca, 0x5e, 0x73, 0x8a, 0xb3, 0xaf, 0x00, 0x00, 0x00, 0xff, 0xff, 0x9f, 0xc4, 0x7f,
	0xf4, 0x98, 0x01, 0x00, 0x00,
}
