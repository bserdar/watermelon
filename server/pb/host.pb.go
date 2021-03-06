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
	// 284 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x51, 0xc1, 0x4a, 0xc3, 0x40,
	0x10, 0x25, 0x49, 0x5b, 0xed, 0x54, 0x54, 0x16, 0x91, 0x45, 0x3c, 0x84, 0x5c, 0x4c, 0x2f, 0x09,
	0xa8, 0xa0, 0x88, 0x1e, 0xd4, 0x16, 0x0c, 0xa8, 0x48, 0x8e, 0xde, 0x76, 0xcd, 0xd8, 0x16, 0x93,
	0x6c, 0x98, 0xdd, 0x56, 0xfa, 0xab, 0x7e, 0x8d, 0x6c, 0xdc, 0x58, 0xf1, 0xf6, 0xde, 0xbc, 0x99,
	0x37, 0xbb, 0x6f, 0x00, 0xe6, 0x4a, 0x9b, 0xa4, 0x21, 0x65, 0x14, 0xf3, 0x1b, 0x19, 0x5d, 0xc0,
	0xd6, 0x6d, 0x51, 0x10, 0x6a, 0xcd, 0x18, 0xf4, 0x9e, 0x45, 0x85, 0xdc, 0x0b, 0xbd, 0x78, 0x98,
	0xb7, 0x98, 0xf1, 0x5f, 0x99, 0xfb, 0x6d, 0xb9, 0xa3, 0xd1, 0x97, 0x07, 0xdb, 0x0f, 0x4a, 0x9b,
	0xac, 0x7e, 0x57, 0x6c, 0x17, 0xfc, 0x6c, 0xe2, 0x06, 0xfd, 0x6c, 0xc2, 0xc6, 0x30, 0x74, 0x7d,
	0xa8, 0x79, 0x10, 0x06, 0xf1, 0xe8, 0x74, 0x94, 0x34, 0x32, 0x71, 0xc5, 0x7c, 0xa3, 0xb2, 0x43,
	0x18, 0x3c, 0x0a, 0x89, 0xa5, 0xe6, 0xbd, 0x30, 0x88, 0x87, 0xb9, 0x63, 0xec, 0x1a, 0xe0, 0x85,
	0x54, 0x83, 0x64, 0x16, 0xa8, 0x79, 0xbf, 0xf5, 0x38, 0xb6, 0x1e, 0xdd, 0xd2, 0x64, 0x23, 0x4f,
	0x6b, 0x43, 0xeb, 0xfc, 0x4f, 0xff, 0xd1, 0x0d, 0xec, 0xfd, 0x93, 0xd9, 0x3e, 0x04, 0x1f, 0xb8,
	0x76, 0x8f, 0xb4, 0x90, 0x1d, 0x40, 0x7f, 0x25, 0xca, 0x25, 0xba, 0xaf, 0xfd, 0x90, 0x2b, 0xff,
	0xd2, 0x8b, 0xce, 0x61, 0xe7, 0x5e, 0x55, 0x95, 0xa8, 0x8b, 0x29, 0x91, 0x22, 0x1b, 0x8d, 0x5d,
	0xdb, 0x45, 0x63, 0xb1, 0xf5, 0x7b, 0xd2, 0x33, 0x37, 0x6b, 0xe1, 0xdd, 0xf8, 0xf5, 0x64, 0xb6,
	0x30, 0xf3, 0xa5, 0x4c, 0xde, 0x54, 0x95, 0x4a, 0x8d, 0x54, 0x08, 0x4a, 0x3f, 0x85, 0x41, 0xaa,
	0xb0, 0x54, 0x75, 0xaa, 0x91, 0x56, 0x48, 0x69, 0x23, 0xe5, 0xa0, 0xbd, 0xc0, 0xd9, 0x77, 0x00,
	0x00, 0x00, 0xff, 0xff, 0x67, 0x86, 0xef, 0x68, 0x8f, 0x01, 0x00, 0x00,
}
