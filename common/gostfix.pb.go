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
	PlainText            string              `protobuf:"bytes,1,opt,name=plainText,proto3" json:"plainText,omitempty"`
	RichText             string              `protobuf:"bytes,2,opt,name=richText,proto3" json:"richText,omitempty"`
	Attachments          []*AttachmentHeader `protobuf:"bytes,3,rep,name=attachments,proto3" json:"attachments,omitempty"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-" bson:"-"`
	XXX_unrecognized     []byte              `json:"-" bson:"-"`
	XXX_sizecache        int32               `json:"-" bson:"-"`
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

func (m *MailBody) GetPlainText() string {
	if m != nil {
		return m.PlainText
	}
	return ""
}

func (m *MailBody) GetRichText() string {
	if m != nil {
		return m.RichText
	}
	return ""
}

func (m *MailBody) GetAttachments() []*AttachmentHeader {
	if m != nil {
		return m.Attachments
	}
	return nil
}

type MailHeader struct {
	From                 string   `protobuf:"bytes,1,opt,name=from,proto3" json:"from,omitempty"`
	To                   string   `protobuf:"bytes,2,opt,name=to,proto3" json:"to,omitempty"`
	Cc                   string   `protobuf:"bytes,3,opt,name=cc,proto3" json:"cc,omitempty"`
	Bcc                  string   `protobuf:"bytes,4,opt,name=bcc,proto3" json:"bcc,omitempty"`
	Date                 int64    `protobuf:"zigzag64,5,opt,name=date,proto3" json:"date,omitempty"`
	Subject              string   `protobuf:"bytes,6,opt,name=subject,proto3" json:"subject,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" bson:"-"`
	XXX_unrecognized     []byte   `json:"-" bson:"-"`
	XXX_sizecache        int32    `json:"-" bson:"-"`
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

func (m *MailHeader) GetDate() int64 {
	if m != nil {
		return m.Date
	}
	return 0
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
	XXX_NoUnkeyedLiteral struct{}    `json:"-" bson:"-"`
	XXX_unrecognized     []byte      `json:"-" bson:"-"`
	XXX_sizecache        int32       `json:"-" bson:"-"`
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

type Attachment struct {
	Header               *AttachmentHeader `protobuf:"bytes,1,opt,name=header,proto3" json:"header,omitempty"`
	Data                 []byte            `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-" bson:"-"`
	XXX_unrecognized     []byte            `json:"-" bson:"-"`
	XXX_sizecache        int32             `json:"-" bson:"-"`
}

func (m *Attachment) Reset()         { *m = Attachment{} }
func (m *Attachment) String() string { return proto.CompactTextString(m) }
func (*Attachment) ProtoMessage()    {}
func (*Attachment) Descriptor() ([]byte, []int) {
	return fileDescriptor_0ab36b6dc6e1dcaa, []int{3}
}

func (m *Attachment) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Attachment.Unmarshal(m, b)
}
func (m *Attachment) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Attachment.Marshal(b, m, deterministic)
}
func (m *Attachment) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Attachment.Merge(m, src)
}
func (m *Attachment) XXX_Size() int {
	return xxx_messageInfo_Attachment.Size(m)
}
func (m *Attachment) XXX_DiscardUnknown() {
	xxx_messageInfo_Attachment.DiscardUnknown(m)
}

var xxx_messageInfo_Attachment proto.InternalMessageInfo

func (m *Attachment) GetHeader() *AttachmentHeader {
	if m != nil {
		return m.Header
	}
	return nil
}

func (m *Attachment) GetData() []byte {
	if m != nil {
		return m.Data
	}
	return nil
}

type AttachmentHeader struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	FileName             string   `protobuf:"bytes,2,opt,name=fileName,proto3" json:"fileName,omitempty"`
	ContentType          string   `protobuf:"bytes,3,opt,name=contentType,proto3" json:"contentType,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" bson:"-"`
	XXX_unrecognized     []byte   `json:"-" bson:"-"`
	XXX_sizecache        int32    `json:"-" bson:"-"`
}

func (m *AttachmentHeader) Reset()         { *m = AttachmentHeader{} }
func (m *AttachmentHeader) String() string { return proto.CompactTextString(m) }
func (*AttachmentHeader) ProtoMessage()    {}
func (*AttachmentHeader) Descriptor() ([]byte, []int) {
	return fileDescriptor_0ab36b6dc6e1dcaa, []int{4}
}

func (m *AttachmentHeader) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AttachmentHeader.Unmarshal(m, b)
}
func (m *AttachmentHeader) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AttachmentHeader.Marshal(b, m, deterministic)
}
func (m *AttachmentHeader) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AttachmentHeader.Merge(m, src)
}
func (m *AttachmentHeader) XXX_Size() int {
	return xxx_messageInfo_AttachmentHeader.Size(m)
}
func (m *AttachmentHeader) XXX_DiscardUnknown() {
	xxx_messageInfo_AttachmentHeader.DiscardUnknown(m)
}

var xxx_messageInfo_AttachmentHeader proto.InternalMessageInfo

func (m *AttachmentHeader) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *AttachmentHeader) GetFileName() string {
	if m != nil {
		return m.FileName
	}
	return ""
}

func (m *AttachmentHeader) GetContentType() string {
	if m != nil {
		return m.ContentType
	}
	return ""
}

type UserInfo struct {
	User                 string   `protobuf:"bytes,1,opt,name=user,proto3" json:"user,omitempty"`
	FullName             string   `protobuf:"bytes,2,opt,name=fullName,proto3" json:"fullName,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" bson:"-"`
	XXX_unrecognized     []byte   `json:"-" bson:"-"`
	XXX_sizecache        int32    `json:"-" bson:"-"`
}

func (m *UserInfo) Reset()         { *m = UserInfo{} }
func (m *UserInfo) String() string { return proto.CompactTextString(m) }
func (*UserInfo) ProtoMessage()    {}
func (*UserInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_0ab36b6dc6e1dcaa, []int{5}
}

func (m *UserInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UserInfo.Unmarshal(m, b)
}
func (m *UserInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UserInfo.Marshal(b, m, deterministic)
}
func (m *UserInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UserInfo.Merge(m, src)
}
func (m *UserInfo) XXX_Size() int {
	return xxx_messageInfo_UserInfo.Size(m)
}
func (m *UserInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_UserInfo.DiscardUnknown(m)
}

var xxx_messageInfo_UserInfo proto.InternalMessageInfo

func (m *UserInfo) GetUser() string {
	if m != nil {
		return m.User
	}
	return ""
}

func (m *UserInfo) GetFullName() string {
	if m != nil {
		return m.FullName
	}
	return ""
}

type Frame struct {
	Skip                 int32    `protobuf:"zigzag32,1,opt,name=skip,proto3" json:"skip,omitempty"`
	Limit                int32    `protobuf:"zigzag32,2,opt,name=limit,proto3" json:"limit,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" bson:"-"`
	XXX_unrecognized     []byte   `json:"-" bson:"-"`
	XXX_sizecache        int32    `json:"-" bson:"-"`
}

func (m *Frame) Reset()         { *m = Frame{} }
func (m *Frame) String() string { return proto.CompactTextString(m) }
func (*Frame) ProtoMessage()    {}
func (*Frame) Descriptor() ([]byte, []int) {
	return fileDescriptor_0ab36b6dc6e1dcaa, []int{6}
}

func (m *Frame) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Frame.Unmarshal(m, b)
}
func (m *Frame) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Frame.Marshal(b, m, deterministic)
}
func (m *Frame) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Frame.Merge(m, src)
}
func (m *Frame) XXX_Size() int {
	return xxx_messageInfo_Frame.Size(m)
}
func (m *Frame) XXX_DiscardUnknown() {
	xxx_messageInfo_Frame.DiscardUnknown(m)
}

var xxx_messageInfo_Frame proto.InternalMessageInfo

func (m *Frame) GetSkip() int32 {
	if m != nil {
		return m.Skip
	}
	return 0
}

func (m *Frame) GetLimit() int32 {
	if m != nil {
		return m.Limit
	}
	return 0
}

type Folder struct {
	Name                 string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Custom               bool     `protobuf:"varint,2,opt,name=custom,proto3" json:"custom,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" bson:"-"`
	XXX_unrecognized     []byte   `json:"-" bson:"-"`
	XXX_sizecache        int32    `json:"-" bson:"-"`
}

func (m *Folder) Reset()         { *m = Folder{} }
func (m *Folder) String() string { return proto.CompactTextString(m) }
func (*Folder) ProtoMessage()    {}
func (*Folder) Descriptor() ([]byte, []int) {
	return fileDescriptor_0ab36b6dc6e1dcaa, []int{7}
}

func (m *Folder) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Folder.Unmarshal(m, b)
}
func (m *Folder) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Folder.Marshal(b, m, deterministic)
}
func (m *Folder) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Folder.Merge(m, src)
}
func (m *Folder) XXX_Size() int {
	return xxx_messageInfo_Folder.Size(m)
}
func (m *Folder) XXX_DiscardUnknown() {
	xxx_messageInfo_Folder.DiscardUnknown(m)
}

var xxx_messageInfo_Folder proto.InternalMessageInfo

func (m *Folder) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *Folder) GetCustom() bool {
	if m != nil {
		return m.Custom
	}
	return false
}

func init() {
	proto.RegisterType((*MailBody)(nil), "common.MailBody")
	proto.RegisterType((*MailHeader)(nil), "common.MailHeader")
	proto.RegisterType((*Mail)(nil), "common.Mail")
	proto.RegisterType((*Attachment)(nil), "common.Attachment")
	proto.RegisterType((*AttachmentHeader)(nil), "common.AttachmentHeader")
	proto.RegisterType((*UserInfo)(nil), "common.UserInfo")
	proto.RegisterType((*Frame)(nil), "common.Frame")
	proto.RegisterType((*Folder)(nil), "common.Folder")
}

func init() {
	proto.RegisterFile("gostfix.proto", fileDescriptor_0ab36b6dc6e1dcaa)
}

var fileDescriptor_0ab36b6dc6e1dcaa = []byte{
	// 403 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x52, 0x4d, 0x8f, 0xd3, 0x30,
	0x10, 0x55, 0x3e, 0x1a, 0xd2, 0x09, 0xa0, 0xee, 0x08, 0x21, 0x0b, 0x71, 0x88, 0x22, 0x0e, 0x15,
	0x87, 0x0a, 0x0a, 0xa7, 0xbd, 0xc1, 0x61, 0x05, 0x07, 0x38, 0x58, 0x8b, 0xc4, 0x11, 0xc7, 0x71,
	0xa8, 0x21, 0x89, 0xa3, 0xc4, 0x91, 0xb6, 0x37, 0x7e, 0x3a, 0xf2, 0xc4, 0x69, 0x4b, 0x25, 0x6e,
	0xef, 0x79, 0xc6, 0xf3, 0xde, 0x3c, 0x1b, 0x9e, 0xfc, 0x34, 0xa3, 0xad, 0xf5, 0xc3, 0xae, 0x1f,
	0x8c, 0x35, 0x98, 0x48, 0xd3, 0xb6, 0xa6, 0x2b, 0xfe, 0x04, 0x90, 0x7e, 0x11, 0xba, 0xf9, 0x68,
	0xaa, 0x23, 0xbe, 0x84, 0x75, 0xdf, 0x08, 0xdd, 0xdd, 0xab, 0x07, 0xcb, 0x82, 0x3c, 0xd8, 0xae,
	0xf9, 0xf9, 0x00, 0x5f, 0x40, 0x3a, 0x68, 0x79, 0xa0, 0x62, 0x48, 0xc5, 0x13, 0xc7, 0x5b, 0xc8,
	0x84, 0xb5, 0x42, 0x1e, 0x5a, 0xd5, 0xd9, 0x91, 0x45, 0x79, 0xb4, 0xcd, 0xf6, 0x6c, 0x37, 0x8b,
	0xec, 0x3e, 0x9c, 0x4a, 0x9f, 0x94, 0xa8, 0xd4, 0xc0, 0x2f, 0x9b, 0x9d, 0x05, 0x70, 0x16, 0xe6,
	0x1a, 0x22, 0xc4, 0xf5, 0x60, 0x5a, 0xaf, 0x4f, 0x18, 0x9f, 0x42, 0x68, 0x8d, 0x17, 0x0d, 0xad,
	0x71, 0x5c, 0x4a, 0x16, 0xcd, 0x5c, 0x4a, 0xdc, 0x40, 0x54, 0x4a, 0xc9, 0x62, 0x3a, 0x70, 0xd0,
	0x4d, 0xa9, 0x84, 0x55, 0x6c, 0x95, 0x07, 0x5b, 0xe4, 0x84, 0x91, 0xc1, 0xa3, 0x71, 0x2a, 0x7f,
	0x29, 0x69, 0x59, 0x42, 0x9d, 0x0b, 0x2d, 0xbe, 0x43, 0xec, 0x1c, 0xe0, 0x6b, 0x48, 0x0e, 0xe4,
	0x82, 0xd4, 0xb3, 0x3d, 0x2e, 0x1b, 0x9c, 0xfd, 0x71, 0xdf, 0x81, 0xaf, 0x20, 0x2e, 0x4d, 0x75,
	0x24, 0x57, 0xd9, 0x7e, 0x73, 0xd9, 0xe9, 0xc2, 0xe4, 0x54, 0x2d, 0x38, 0xc0, 0x79, 0x7b, 0x7c,
	0x73, 0x35, 0xff, 0xff, 0x09, 0x2d, 0x2a, 0xf3, 0x1e, 0x82, 0x54, 0x1e, 0xd3, 0x1e, 0xa2, 0xf8,
	0x01, 0x9b, 0xeb, 0x7e, 0x97, 0x88, 0xae, 0x7c, 0x66, 0xa1, 0xae, 0xdc, 0x63, 0xd5, 0xba, 0x51,
	0x5f, 0x45, 0xab, 0x96, 0xc7, 0x5a, 0x38, 0xe6, 0x90, 0x49, 0xd3, 0x59, 0xd5, 0xd9, 0xfb, 0x63,
	0xaf, 0x7c, 0x8c, 0x97, 0x47, 0xc5, 0x2d, 0xa4, 0xdf, 0x46, 0x35, 0x7c, 0xee, 0x6a, 0xe3, 0x1c,
	0x4c, 0xa3, 0x77, 0xbc, 0xe6, 0x84, 0x69, 0xfa, 0xd4, 0x34, 0xff, 0x4c, 0xf7, 0xbc, 0x78, 0x0b,
	0xab, 0xbb, 0xc1, 0xc9, 0x20, 0xc4, 0xe3, 0x6f, 0xdd, 0xd3, 0xc5, 0x1b, 0x4e, 0x18, 0x9f, 0xc1,
	0xaa, 0xd1, 0xad, 0x9e, 0x3f, 0xd0, 0x0d, 0x9f, 0x49, 0xf1, 0x1e, 0x92, 0x3b, 0xd3, 0xf8, 0x75,
	0x3b, 0x37, 0xd4, 0x8b, 0x39, 0x8c, 0xcf, 0x21, 0x91, 0xd3, 0x68, 0x4d, 0x4b, 0x97, 0x52, 0xee,
	0x59, 0x99, 0xd0, 0x4f, 0x7e, 0xf7, 0x37, 0x00, 0x00, 0xff, 0xff, 0xe9, 0x88, 0x74, 0xc2, 0xda,
	0x02, 0x00, 0x00,
}
