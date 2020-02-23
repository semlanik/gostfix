// Code generated by protoc-gen-go. DO NOT EDIT.
// source: gostfix.proto

package common

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
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type MailBody struct {
	ContentType          string   `protobuf:"bytes,1,opt,name=contentType,proto3" json:"contentType,omitempty"`
	Content              []byte   `protobuf:"bytes,2,opt,name=content,proto3" json:"content,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MailBody) Reset()         { *m = MailBody{} }
func (m *MailBody) String() string { return proto.CompactTextString(m) }
func (*MailBody) ProtoMessage()    {}
func (*MailBody) Descriptor() ([]byte, []int) {
	return fileDescriptor_0ab36b6dc6e1dcaa, []int{0}
}

func (m *MailBody) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MailBody.Unmarshal(m, b)
}
func (m *MailBody) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MailBody.Marshal(b, m, deterministic)
}
func (m *MailBody) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MailBody.Merge(m, src)
}
func (m *MailBody) XXX_Size() int {
	return xxx_messageInfo_MailBody.Size(m)
}
func (m *MailBody) XXX_DiscardUnknown() {
	xxx_messageInfo_MailBody.DiscardUnknown(m)
}

var xxx_messageInfo_MailBody proto.InternalMessageInfo

func (m *MailBody) GetContentType() string {
	if m != nil {
		return m.ContentType
	}
	return ""
}

func (m *MailBody) GetContent() []byte {
	if m != nil {
		return m.Content
	}
	return nil
}

type MailHeader struct {
	From                 string   `protobuf:"bytes,1,opt,name=from,proto3" json:"from,omitempty"`
	To                   string   `protobuf:"bytes,2,opt,name=to,proto3" json:"to,omitempty"`
	Cc                   string   `protobuf:"bytes,3,opt,name=cc,proto3" json:"cc,omitempty"`
	Bcc                  string   `protobuf:"bytes,4,opt,name=bcc,proto3" json:"bcc,omitempty"`
	Date                 string   `protobuf:"bytes,5,opt,name=date,proto3" json:"date,omitempty"`
	Subject              string   `protobuf:"bytes,6,opt,name=subject,proto3" json:"subject,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MailHeader) Reset()         { *m = MailHeader{} }
func (m *MailHeader) String() string { return proto.CompactTextString(m) }
func (*MailHeader) ProtoMessage()    {}
func (*MailHeader) Descriptor() ([]byte, []int) {
	return fileDescriptor_0ab36b6dc6e1dcaa, []int{1}
}

func (m *MailHeader) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MailHeader.Unmarshal(m, b)
}
func (m *MailHeader) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MailHeader.Marshal(b, m, deterministic)
}
func (m *MailHeader) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MailHeader.Merge(m, src)
}
func (m *MailHeader) XXX_Size() int {
	return xxx_messageInfo_MailHeader.Size(m)
}
func (m *MailHeader) XXX_DiscardUnknown() {
	xxx_messageInfo_MailHeader.DiscardUnknown(m)
}

var xxx_messageInfo_MailHeader proto.InternalMessageInfo

func (m *MailHeader) GetFrom() string {
	if m != nil {
		return m.From
	}
	return ""
}

func (m *MailHeader) GetTo() string {
	if m != nil {
		return m.To
	}
	return ""
}

func (m *MailHeader) GetCc() string {
	if m != nil {
		return m.Cc
	}
	return ""
}

func (m *MailHeader) GetBcc() string {
	if m != nil {
		return m.Bcc
	}
	return ""
}

func (m *MailHeader) GetDate() string {
	if m != nil {
		return m.Date
	}
	return ""
}

func (m *MailHeader) GetSubject() string {
	if m != nil {
		return m.Subject
	}
	return ""
}

type Mail struct {
	Header               *MailHeader `protobuf:"bytes,1,opt,name=header,proto3" json:"header,omitempty"`
	Body                 *MailBody   `protobuf:"bytes,2,opt,name=body,proto3" json:"body,omitempty"`
	XXX_NoUnkeyedLiteral struct{}    `json:"-"`
	XXX_unrecognized     []byte      `json:"-"`
	XXX_sizecache        int32       `json:"-"`
}

func (m *Mail) Reset()         { *m = Mail{} }
func (m *Mail) String() string { return proto.CompactTextString(m) }
func (*Mail) ProtoMessage()    {}
func (*Mail) Descriptor() ([]byte, []int) {
	return fileDescriptor_0ab36b6dc6e1dcaa, []int{2}
}

func (m *Mail) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Mail.Unmarshal(m, b)
}
func (m *Mail) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Mail.Marshal(b, m, deterministic)
}
func (m *Mail) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Mail.Merge(m, src)
}
func (m *Mail) XXX_Size() int {
	return xxx_messageInfo_Mail.Size(m)
}
func (m *Mail) XXX_DiscardUnknown() {
	xxx_messageInfo_Mail.DiscardUnknown(m)
}

var xxx_messageInfo_Mail proto.InternalMessageInfo

func (m *Mail) GetHeader() *MailHeader {
	if m != nil {
		return m.Header
	}
	return nil
}

func (m *Mail) GetBody() *MailBody {
	if m != nil {
		return m.Body
	}
	return nil
}

func init() {
	proto.RegisterType((*MailBody)(nil), "common.MailBody")
	proto.RegisterType((*MailHeader)(nil), "common.MailHeader")
	proto.RegisterType((*Mail)(nil), "common.Mail")
}

func init() { proto.RegisterFile("gostfix.proto", fileDescriptor_0ab36b6dc6e1dcaa) }

var fileDescriptor_0ab36b6dc6e1dcaa = []byte{
	// 230 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x4c, 0x90, 0x31, 0x4f, 0xc3, 0x40,
	0x0c, 0x85, 0x95, 0x34, 0x04, 0xea, 0x00, 0xaa, 0x3c, 0xdd, 0x18, 0x45, 0x0c, 0x15, 0x43, 0x86,
	0xf2, 0x0f, 0x18, 0x10, 0x0b, 0xcb, 0x89, 0x81, 0x35, 0xf1, 0x5d, 0xa1, 0x88, 0xc4, 0x55, 0x6a,
	0x24, 0xb2, 0xf1, 0xd3, 0x91, 0x9d, 0x54, 0x64, 0x7b, 0xef, 0x9d, 0xef, 0xdd, 0x77, 0x86, 0x9b,
	0x77, 0x3e, 0xc9, 0xfe, 0xf0, 0x53, 0x1f, 0x07, 0x16, 0xc6, 0x9c, 0xb8, 0xeb, 0xb8, 0xaf, 0x9e,
	0xe0, 0xea, 0xa5, 0x39, 0x7c, 0x3d, 0x72, 0x18, 0xb1, 0x84, 0x82, 0xb8, 0x97, 0xd8, 0xcb, 0xeb,
	0x78, 0x8c, 0x2e, 0x29, 0x93, 0xed, 0xda, 0x2f, 0x23, 0x74, 0x70, 0x39, 0x5b, 0x97, 0x96, 0xc9,
	0xf6, 0xda, 0x9f, 0x6d, 0xf5, 0x9b, 0x00, 0x68, 0xd1, 0x73, 0x6c, 0x42, 0x1c, 0x10, 0x21, 0xdb,
	0x0f, 0xdc, 0xcd, 0x1d, 0xa6, 0xf1, 0x16, 0x52, 0x61, 0xbb, 0xb7, 0xf6, 0xa9, 0xb0, 0x7a, 0x22,
	0xb7, 0x9a, 0x3c, 0x11, 0x6e, 0x60, 0xd5, 0x12, 0xb9, 0xcc, 0x02, 0x95, 0xda, 0x12, 0x1a, 0x89,
	0xee, 0x62, 0x6a, 0x51, 0xad, 0x08, 0xa7, 0xef, 0xf6, 0x33, 0x92, 0xb8, 0xdc, 0xe2, 0xb3, 0xad,
	0xde, 0x20, 0x53, 0x02, 0xbc, 0x87, 0xfc, 0xc3, 0x28, 0xec, 0xf5, 0x62, 0x87, 0xf5, 0xf4, 0xd7,
	0xfa, 0x9f, 0xcf, 0xcf, 0x13, 0x78, 0x07, 0x59, 0xcb, 0x61, 0x34, 0xaa, 0x62, 0xb7, 0x59, 0x4e,
	0xea, 0x4a, 0xbc, 0x9d, 0xb6, 0xb9, 0xed, 0xec, 0xe1, 0x2f, 0x00, 0x00, 0xff, 0xff, 0x81, 0xe6,
	0x8d, 0x59, 0x44, 0x01, 0x00, 0x00,
}
