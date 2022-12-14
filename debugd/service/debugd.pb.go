// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.8
// source: debugd.proto

package service

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type UploadBootstrapperStatus int32

const (
	UploadBootstrapperStatus_UPLOAD_BOOTSTRAPPER_SUCCESS       UploadBootstrapperStatus = 0
	UploadBootstrapperStatus_UPLOAD_BOOTSTRAPPER_UPLOAD_FAILED UploadBootstrapperStatus = 1
	UploadBootstrapperStatus_UPLOAD_BOOTSTRAPPER_START_FAILED  UploadBootstrapperStatus = 2
	UploadBootstrapperStatus_UPLOAD_BOOTSTRAPPER_FILE_EXISTS   UploadBootstrapperStatus = 3
)

// Enum value maps for UploadBootstrapperStatus.
var (
	UploadBootstrapperStatus_name = map[int32]string{
		0: "UPLOAD_BOOTSTRAPPER_SUCCESS",
		1: "UPLOAD_BOOTSTRAPPER_UPLOAD_FAILED",
		2: "UPLOAD_BOOTSTRAPPER_START_FAILED",
		3: "UPLOAD_BOOTSTRAPPER_FILE_EXISTS",
	}
	UploadBootstrapperStatus_value = map[string]int32{
		"UPLOAD_BOOTSTRAPPER_SUCCESS":       0,
		"UPLOAD_BOOTSTRAPPER_UPLOAD_FAILED": 1,
		"UPLOAD_BOOTSTRAPPER_START_FAILED":  2,
		"UPLOAD_BOOTSTRAPPER_FILE_EXISTS":   3,
	}
)

func (x UploadBootstrapperStatus) Enum() *UploadBootstrapperStatus {
	p := new(UploadBootstrapperStatus)
	*p = x
	return p
}

func (x UploadBootstrapperStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (UploadBootstrapperStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_debugd_proto_enumTypes[0].Descriptor()
}

func (UploadBootstrapperStatus) Type() protoreflect.EnumType {
	return &file_debugd_proto_enumTypes[0]
}

func (x UploadBootstrapperStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use UploadBootstrapperStatus.Descriptor instead.
func (UploadBootstrapperStatus) EnumDescriptor() ([]byte, []int) {
	return file_debugd_proto_rawDescGZIP(), []int{0}
}

type UploadSystemdServiceUnitsStatus int32

const (
	UploadSystemdServiceUnitsStatus_UPLOAD_SYSTEMD_SERVICE_UNITS_SUCCESS UploadSystemdServiceUnitsStatus = 0
	UploadSystemdServiceUnitsStatus_UPLOAD_SYSTEMD_SERVICE_UNITS_FAILURE UploadSystemdServiceUnitsStatus = 1
)

// Enum value maps for UploadSystemdServiceUnitsStatus.
var (
	UploadSystemdServiceUnitsStatus_name = map[int32]string{
		0: "UPLOAD_SYSTEMD_SERVICE_UNITS_SUCCESS",
		1: "UPLOAD_SYSTEMD_SERVICE_UNITS_FAILURE",
	}
	UploadSystemdServiceUnitsStatus_value = map[string]int32{
		"UPLOAD_SYSTEMD_SERVICE_UNITS_SUCCESS": 0,
		"UPLOAD_SYSTEMD_SERVICE_UNITS_FAILURE": 1,
	}
)

func (x UploadSystemdServiceUnitsStatus) Enum() *UploadSystemdServiceUnitsStatus {
	p := new(UploadSystemdServiceUnitsStatus)
	*p = x
	return p
}

func (x UploadSystemdServiceUnitsStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (UploadSystemdServiceUnitsStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_debugd_proto_enumTypes[1].Descriptor()
}

func (UploadSystemdServiceUnitsStatus) Type() protoreflect.EnumType {
	return &file_debugd_proto_enumTypes[1]
}

func (x UploadSystemdServiceUnitsStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use UploadSystemdServiceUnitsStatus.Descriptor instead.
func (UploadSystemdServiceUnitsStatus) EnumDescriptor() ([]byte, []int) {
	return file_debugd_proto_rawDescGZIP(), []int{1}
}

type SetInfoRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Info []*Info `protobuf:"bytes,1,rep,name=info,proto3" json:"info,omitempty"`
}

func (x *SetInfoRequest) Reset() {
	*x = SetInfoRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_debugd_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetInfoRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetInfoRequest) ProtoMessage() {}

func (x *SetInfoRequest) ProtoReflect() protoreflect.Message {
	mi := &file_debugd_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetInfoRequest.ProtoReflect.Descriptor instead.
func (*SetInfoRequest) Descriptor() ([]byte, []int) {
	return file_debugd_proto_rawDescGZIP(), []int{0}
}

func (x *SetInfoRequest) GetInfo() []*Info {
	if x != nil {
		return x.Info
	}
	return nil
}

type SetInfoResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *SetInfoResponse) Reset() {
	*x = SetInfoResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_debugd_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetInfoResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetInfoResponse) ProtoMessage() {}

func (x *SetInfoResponse) ProtoReflect() protoreflect.Message {
	mi := &file_debugd_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetInfoResponse.ProtoReflect.Descriptor instead.
func (*SetInfoResponse) Descriptor() ([]byte, []int) {
	return file_debugd_proto_rawDescGZIP(), []int{1}
}

type GetInfoRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *GetInfoRequest) Reset() {
	*x = GetInfoRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_debugd_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetInfoRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetInfoRequest) ProtoMessage() {}

func (x *GetInfoRequest) ProtoReflect() protoreflect.Message {
	mi := &file_debugd_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetInfoRequest.ProtoReflect.Descriptor instead.
func (*GetInfoRequest) Descriptor() ([]byte, []int) {
	return file_debugd_proto_rawDescGZIP(), []int{2}
}

type GetInfoResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Info []*Info `protobuf:"bytes,1,rep,name=info,proto3" json:"info,omitempty"`
}

func (x *GetInfoResponse) Reset() {
	*x = GetInfoResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_debugd_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetInfoResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetInfoResponse) ProtoMessage() {}

func (x *GetInfoResponse) ProtoReflect() protoreflect.Message {
	mi := &file_debugd_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetInfoResponse.ProtoReflect.Descriptor instead.
func (*GetInfoResponse) Descriptor() ([]byte, []int) {
	return file_debugd_proto_rawDescGZIP(), []int{3}
}

func (x *GetInfoResponse) GetInfo() []*Info {
	if x != nil {
		return x.Info
	}
	return nil
}

type Info struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key   string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value string `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *Info) Reset() {
	*x = Info{}
	if protoimpl.UnsafeEnabled {
		mi := &file_debugd_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Info) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Info) ProtoMessage() {}

func (x *Info) ProtoReflect() protoreflect.Message {
	mi := &file_debugd_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Info.ProtoReflect.Descriptor instead.
func (*Info) Descriptor() ([]byte, []int) {
	return file_debugd_proto_rawDescGZIP(), []int{4}
}

func (x *Info) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *Info) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

type DownloadBootstrapperRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *DownloadBootstrapperRequest) Reset() {
	*x = DownloadBootstrapperRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_debugd_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DownloadBootstrapperRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DownloadBootstrapperRequest) ProtoMessage() {}

func (x *DownloadBootstrapperRequest) ProtoReflect() protoreflect.Message {
	mi := &file_debugd_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DownloadBootstrapperRequest.ProtoReflect.Descriptor instead.
func (*DownloadBootstrapperRequest) Descriptor() ([]byte, []int) {
	return file_debugd_proto_rawDescGZIP(), []int{5}
}

type Chunk struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Content []byte `protobuf:"bytes,1,opt,name=content,proto3" json:"content,omitempty"`
}

func (x *Chunk) Reset() {
	*x = Chunk{}
	if protoimpl.UnsafeEnabled {
		mi := &file_debugd_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Chunk) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Chunk) ProtoMessage() {}

func (x *Chunk) ProtoReflect() protoreflect.Message {
	mi := &file_debugd_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Chunk.ProtoReflect.Descriptor instead.
func (*Chunk) Descriptor() ([]byte, []int) {
	return file_debugd_proto_rawDescGZIP(), []int{6}
}

func (x *Chunk) GetContent() []byte {
	if x != nil {
		return x.Content
	}
	return nil
}

type UploadBootstrapperResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status UploadBootstrapperStatus `protobuf:"varint,1,opt,name=status,proto3,enum=debugd.UploadBootstrapperStatus" json:"status,omitempty"`
}

func (x *UploadBootstrapperResponse) Reset() {
	*x = UploadBootstrapperResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_debugd_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UploadBootstrapperResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UploadBootstrapperResponse) ProtoMessage() {}

func (x *UploadBootstrapperResponse) ProtoReflect() protoreflect.Message {
	mi := &file_debugd_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UploadBootstrapperResponse.ProtoReflect.Descriptor instead.
func (*UploadBootstrapperResponse) Descriptor() ([]byte, []int) {
	return file_debugd_proto_rawDescGZIP(), []int{7}
}

func (x *UploadBootstrapperResponse) GetStatus() UploadBootstrapperStatus {
	if x != nil {
		return x.Status
	}
	return UploadBootstrapperStatus_UPLOAD_BOOTSTRAPPER_SUCCESS
}

type ServiceUnit struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name     string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Contents string `protobuf:"bytes,2,opt,name=contents,proto3" json:"contents,omitempty"`
}

func (x *ServiceUnit) Reset() {
	*x = ServiceUnit{}
	if protoimpl.UnsafeEnabled {
		mi := &file_debugd_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServiceUnit) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServiceUnit) ProtoMessage() {}

func (x *ServiceUnit) ProtoReflect() protoreflect.Message {
	mi := &file_debugd_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ServiceUnit.ProtoReflect.Descriptor instead.
func (*ServiceUnit) Descriptor() ([]byte, []int) {
	return file_debugd_proto_rawDescGZIP(), []int{8}
}

func (x *ServiceUnit) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *ServiceUnit) GetContents() string {
	if x != nil {
		return x.Contents
	}
	return ""
}

type UploadSystemdServiceUnitsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Units []*ServiceUnit `protobuf:"bytes,1,rep,name=units,proto3" json:"units,omitempty"`
}

func (x *UploadSystemdServiceUnitsRequest) Reset() {
	*x = UploadSystemdServiceUnitsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_debugd_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UploadSystemdServiceUnitsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UploadSystemdServiceUnitsRequest) ProtoMessage() {}

func (x *UploadSystemdServiceUnitsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_debugd_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UploadSystemdServiceUnitsRequest.ProtoReflect.Descriptor instead.
func (*UploadSystemdServiceUnitsRequest) Descriptor() ([]byte, []int) {
	return file_debugd_proto_rawDescGZIP(), []int{9}
}

func (x *UploadSystemdServiceUnitsRequest) GetUnits() []*ServiceUnit {
	if x != nil {
		return x.Units
	}
	return nil
}

type UploadSystemdServiceUnitsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Status UploadSystemdServiceUnitsStatus `protobuf:"varint,1,opt,name=status,proto3,enum=debugd.UploadSystemdServiceUnitsStatus" json:"status,omitempty"`
}

func (x *UploadSystemdServiceUnitsResponse) Reset() {
	*x = UploadSystemdServiceUnitsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_debugd_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UploadSystemdServiceUnitsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UploadSystemdServiceUnitsResponse) ProtoMessage() {}

func (x *UploadSystemdServiceUnitsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_debugd_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UploadSystemdServiceUnitsResponse.ProtoReflect.Descriptor instead.
func (*UploadSystemdServiceUnitsResponse) Descriptor() ([]byte, []int) {
	return file_debugd_proto_rawDescGZIP(), []int{10}
}

func (x *UploadSystemdServiceUnitsResponse) GetStatus() UploadSystemdServiceUnitsStatus {
	if x != nil {
		return x.Status
	}
	return UploadSystemdServiceUnitsStatus_UPLOAD_SYSTEMD_SERVICE_UNITS_SUCCESS
}

var File_debugd_proto protoreflect.FileDescriptor

var file_debugd_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x64, 0x65, 0x62, 0x75, 0x67, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06,
	0x64, 0x65, 0x62, 0x75, 0x67, 0x64, 0x22, 0x32, 0x0a, 0x0e, 0x53, 0x65, 0x74, 0x49, 0x6e, 0x66,
	0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x20, 0x0a, 0x04, 0x69, 0x6e, 0x66, 0x6f,
	0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0c, 0x2e, 0x64, 0x65, 0x62, 0x75, 0x67, 0x64, 0x2e,
	0x49, 0x6e, 0x66, 0x6f, 0x52, 0x04, 0x69, 0x6e, 0x66, 0x6f, 0x22, 0x11, 0x0a, 0x0f, 0x53, 0x65,
	0x74, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x10, 0x0a,
	0x0e, 0x47, 0x65, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22,
	0x33, 0x0a, 0x0f, 0x47, 0x65, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x20, 0x0a, 0x04, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x0c, 0x2e, 0x64, 0x65, 0x62, 0x75, 0x67, 0x64, 0x2e, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x04,
	0x69, 0x6e, 0x66, 0x6f, 0x22, 0x2e, 0x0a, 0x04, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x10, 0x0a, 0x03,
	0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14,
	0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x22, 0x1d, 0x0a, 0x1b, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64,
	0x42, 0x6f, 0x6f, 0x74, 0x73, 0x74, 0x72, 0x61, 0x70, 0x70, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x22, 0x21, 0x0a, 0x05, 0x43, 0x68, 0x75, 0x6e, 0x6b, 0x12, 0x18, 0x0a, 0x07,
	0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x63,
	0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x22, 0x56, 0x0a, 0x1a, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64,
	0x42, 0x6f, 0x6f, 0x74, 0x73, 0x74, 0x72, 0x61, 0x70, 0x70, 0x65, 0x72, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x38, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x0e, 0x32, 0x20, 0x2e, 0x64, 0x65, 0x62, 0x75, 0x67, 0x64, 0x2e, 0x55, 0x70,
	0x6c, 0x6f, 0x61, 0x64, 0x42, 0x6f, 0x6f, 0x74, 0x73, 0x74, 0x72, 0x61, 0x70, 0x70, 0x65, 0x72,
	0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x3d,
	0x0a, 0x0b, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x55, 0x6e, 0x69, 0x74, 0x12, 0x12, 0x0a,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d,
	0x65, 0x12, 0x1a, 0x0a, 0x08, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x08, 0x63, 0x6f, 0x6e, 0x74, 0x65, 0x6e, 0x74, 0x73, 0x22, 0x4d, 0x0a,
	0x20, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x64, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x55, 0x6e, 0x69, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x29, 0x0a, 0x05, 0x75, 0x6e, 0x69, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x13, 0x2e, 0x64, 0x65, 0x62, 0x75, 0x67, 0x64, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x55, 0x6e, 0x69, 0x74, 0x52, 0x05, 0x75, 0x6e, 0x69, 0x74, 0x73, 0x22, 0x64, 0x0a, 0x21,
	0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x64, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x55, 0x6e, 0x69, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x12, 0x3f, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x27, 0x2e, 0x64, 0x65, 0x62, 0x75, 0x67, 0x64, 0x2e, 0x55, 0x70, 0x6c, 0x6f, 0x61,
	0x64, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x64, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x55,
	0x6e, 0x69, 0x74, 0x73, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x2a, 0xad, 0x01, 0x0a, 0x18, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x42, 0x6f, 0x6f,
	0x74, 0x73, 0x74, 0x72, 0x61, 0x70, 0x70, 0x65, 0x72, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12,
	0x1f, 0x0a, 0x1b, 0x55, 0x50, 0x4c, 0x4f, 0x41, 0x44, 0x5f, 0x42, 0x4f, 0x4f, 0x54, 0x53, 0x54,
	0x52, 0x41, 0x50, 0x50, 0x45, 0x52, 0x5f, 0x53, 0x55, 0x43, 0x43, 0x45, 0x53, 0x53, 0x10, 0x00,
	0x12, 0x25, 0x0a, 0x21, 0x55, 0x50, 0x4c, 0x4f, 0x41, 0x44, 0x5f, 0x42, 0x4f, 0x4f, 0x54, 0x53,
	0x54, 0x52, 0x41, 0x50, 0x50, 0x45, 0x52, 0x5f, 0x55, 0x50, 0x4c, 0x4f, 0x41, 0x44, 0x5f, 0x46,
	0x41, 0x49, 0x4c, 0x45, 0x44, 0x10, 0x01, 0x12, 0x24, 0x0a, 0x20, 0x55, 0x50, 0x4c, 0x4f, 0x41,
	0x44, 0x5f, 0x42, 0x4f, 0x4f, 0x54, 0x53, 0x54, 0x52, 0x41, 0x50, 0x50, 0x45, 0x52, 0x5f, 0x53,
	0x54, 0x41, 0x52, 0x54, 0x5f, 0x46, 0x41, 0x49, 0x4c, 0x45, 0x44, 0x10, 0x02, 0x12, 0x23, 0x0a,
	0x1f, 0x55, 0x50, 0x4c, 0x4f, 0x41, 0x44, 0x5f, 0x42, 0x4f, 0x4f, 0x54, 0x53, 0x54, 0x52, 0x41,
	0x50, 0x50, 0x45, 0x52, 0x5f, 0x46, 0x49, 0x4c, 0x45, 0x5f, 0x45, 0x58, 0x49, 0x53, 0x54, 0x53,
	0x10, 0x03, 0x2a, 0x75, 0x0a, 0x1f, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x53, 0x79, 0x73, 0x74,
	0x65, 0x6d, 0x64, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x55, 0x6e, 0x69, 0x74, 0x73, 0x53,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x28, 0x0a, 0x24, 0x55, 0x50, 0x4c, 0x4f, 0x41, 0x44, 0x5f,
	0x53, 0x59, 0x53, 0x54, 0x45, 0x4d, 0x44, 0x5f, 0x53, 0x45, 0x52, 0x56, 0x49, 0x43, 0x45, 0x5f,
	0x55, 0x4e, 0x49, 0x54, 0x53, 0x5f, 0x53, 0x55, 0x43, 0x43, 0x45, 0x53, 0x53, 0x10, 0x00, 0x12,
	0x28, 0x0a, 0x24, 0x55, 0x50, 0x4c, 0x4f, 0x41, 0x44, 0x5f, 0x53, 0x59, 0x53, 0x54, 0x45, 0x4d,
	0x44, 0x5f, 0x53, 0x45, 0x52, 0x56, 0x49, 0x43, 0x45, 0x5f, 0x55, 0x4e, 0x49, 0x54, 0x53, 0x5f,
	0x46, 0x41, 0x49, 0x4c, 0x55, 0x52, 0x45, 0x10, 0x01, 0x32, 0x94, 0x03, 0x0a, 0x06, 0x44, 0x65,
	0x62, 0x75, 0x67, 0x64, 0x12, 0x3c, 0x0a, 0x07, 0x53, 0x65, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x12,
	0x16, 0x2e, 0x64, 0x65, 0x62, 0x75, 0x67, 0x64, 0x2e, 0x53, 0x65, 0x74, 0x49, 0x6e, 0x66, 0x6f,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x17, 0x2e, 0x64, 0x65, 0x62, 0x75, 0x67, 0x64,
	0x2e, 0x53, 0x65, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x00, 0x12, 0x3c, 0x0a, 0x07, 0x47, 0x65, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x12, 0x16, 0x2e,
	0x64, 0x65, 0x62, 0x75, 0x67, 0x64, 0x2e, 0x47, 0x65, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x17, 0x2e, 0x64, 0x65, 0x62, 0x75, 0x67, 0x64, 0x2e, 0x47,
	0x65, 0x74, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00,
	0x12, 0x4b, 0x0a, 0x12, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x42, 0x6f, 0x6f, 0x74, 0x73, 0x74,
	0x72, 0x61, 0x70, 0x70, 0x65, 0x72, 0x12, 0x0d, 0x2e, 0x64, 0x65, 0x62, 0x75, 0x67, 0x64, 0x2e,
	0x43, 0x68, 0x75, 0x6e, 0x6b, 0x1a, 0x22, 0x2e, 0x64, 0x65, 0x62, 0x75, 0x67, 0x64, 0x2e, 0x55,
	0x70, 0x6c, 0x6f, 0x61, 0x64, 0x42, 0x6f, 0x6f, 0x74, 0x73, 0x74, 0x72, 0x61, 0x70, 0x70, 0x65,
	0x72, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x28, 0x01, 0x12, 0x4e, 0x0a,
	0x14, 0x44, 0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x42, 0x6f, 0x6f, 0x74, 0x73, 0x74, 0x72,
	0x61, 0x70, 0x70, 0x65, 0x72, 0x12, 0x23, 0x2e, 0x64, 0x65, 0x62, 0x75, 0x67, 0x64, 0x2e, 0x44,
	0x6f, 0x77, 0x6e, 0x6c, 0x6f, 0x61, 0x64, 0x42, 0x6f, 0x6f, 0x74, 0x73, 0x74, 0x72, 0x61, 0x70,
	0x70, 0x65, 0x72, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0d, 0x2e, 0x64, 0x65, 0x62,
	0x75, 0x67, 0x64, 0x2e, 0x43, 0x68, 0x75, 0x6e, 0x6b, 0x22, 0x00, 0x30, 0x01, 0x12, 0x71, 0x0a,
	0x18, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x53, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x55, 0x6e, 0x69, 0x74, 0x73, 0x12, 0x28, 0x2e, 0x64, 0x65, 0x62, 0x75,
	0x67, 0x64, 0x2e, 0x55, 0x70, 0x6c, 0x6f, 0x61, 0x64, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x64,
	0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x55, 0x6e, 0x69, 0x74, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x1a, 0x29, 0x2e, 0x64, 0x65, 0x62, 0x75, 0x67, 0x64, 0x2e, 0x55, 0x70, 0x6c,
	0x6f, 0x61, 0x64, 0x53, 0x79, 0x73, 0x74, 0x65, 0x6d, 0x64, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x55, 0x6e, 0x69, 0x74, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00,
	0x42, 0x38, 0x5a, 0x36, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x65,
	0x64, 0x67, 0x65, 0x6c, 0x65, 0x73, 0x73, 0x73, 0x79, 0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x73, 0x74,
	0x65, 0x6c, 0x6c, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x76, 0x32, 0x2f, 0x64, 0x65, 0x62, 0x75,
	0x67, 0x64, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_debugd_proto_rawDescOnce sync.Once
	file_debugd_proto_rawDescData = file_debugd_proto_rawDesc
)

func file_debugd_proto_rawDescGZIP() []byte {
	file_debugd_proto_rawDescOnce.Do(func() {
		file_debugd_proto_rawDescData = protoimpl.X.CompressGZIP(file_debugd_proto_rawDescData)
	})
	return file_debugd_proto_rawDescData
}

var file_debugd_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_debugd_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_debugd_proto_goTypes = []interface{}{
	(UploadBootstrapperStatus)(0),             // 0: debugd.UploadBootstrapperStatus
	(UploadSystemdServiceUnitsStatus)(0),      // 1: debugd.UploadSystemdServiceUnitsStatus
	(*SetInfoRequest)(nil),                    // 2: debugd.SetInfoRequest
	(*SetInfoResponse)(nil),                   // 3: debugd.SetInfoResponse
	(*GetInfoRequest)(nil),                    // 4: debugd.GetInfoRequest
	(*GetInfoResponse)(nil),                   // 5: debugd.GetInfoResponse
	(*Info)(nil),                              // 6: debugd.Info
	(*DownloadBootstrapperRequest)(nil),       // 7: debugd.DownloadBootstrapperRequest
	(*Chunk)(nil),                             // 8: debugd.Chunk
	(*UploadBootstrapperResponse)(nil),        // 9: debugd.UploadBootstrapperResponse
	(*ServiceUnit)(nil),                       // 10: debugd.ServiceUnit
	(*UploadSystemdServiceUnitsRequest)(nil),  // 11: debugd.UploadSystemdServiceUnitsRequest
	(*UploadSystemdServiceUnitsResponse)(nil), // 12: debugd.UploadSystemdServiceUnitsResponse
}
var file_debugd_proto_depIdxs = []int32{
	6,  // 0: debugd.SetInfoRequest.info:type_name -> debugd.Info
	6,  // 1: debugd.GetInfoResponse.info:type_name -> debugd.Info
	0,  // 2: debugd.UploadBootstrapperResponse.status:type_name -> debugd.UploadBootstrapperStatus
	10, // 3: debugd.UploadSystemdServiceUnitsRequest.units:type_name -> debugd.ServiceUnit
	1,  // 4: debugd.UploadSystemdServiceUnitsResponse.status:type_name -> debugd.UploadSystemdServiceUnitsStatus
	2,  // 5: debugd.Debugd.SetInfo:input_type -> debugd.SetInfoRequest
	4,  // 6: debugd.Debugd.GetInfo:input_type -> debugd.GetInfoRequest
	8,  // 7: debugd.Debugd.UploadBootstrapper:input_type -> debugd.Chunk
	7,  // 8: debugd.Debugd.DownloadBootstrapper:input_type -> debugd.DownloadBootstrapperRequest
	11, // 9: debugd.Debugd.UploadSystemServiceUnits:input_type -> debugd.UploadSystemdServiceUnitsRequest
	3,  // 10: debugd.Debugd.SetInfo:output_type -> debugd.SetInfoResponse
	5,  // 11: debugd.Debugd.GetInfo:output_type -> debugd.GetInfoResponse
	9,  // 12: debugd.Debugd.UploadBootstrapper:output_type -> debugd.UploadBootstrapperResponse
	8,  // 13: debugd.Debugd.DownloadBootstrapper:output_type -> debugd.Chunk
	12, // 14: debugd.Debugd.UploadSystemServiceUnits:output_type -> debugd.UploadSystemdServiceUnitsResponse
	10, // [10:15] is the sub-list for method output_type
	5,  // [5:10] is the sub-list for method input_type
	5,  // [5:5] is the sub-list for extension type_name
	5,  // [5:5] is the sub-list for extension extendee
	0,  // [0:5] is the sub-list for field type_name
}

func init() { file_debugd_proto_init() }
func file_debugd_proto_init() {
	if File_debugd_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_debugd_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetInfoRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_debugd_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetInfoResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_debugd_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetInfoRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_debugd_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetInfoResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_debugd_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Info); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_debugd_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DownloadBootstrapperRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_debugd_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Chunk); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_debugd_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UploadBootstrapperResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_debugd_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServiceUnit); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_debugd_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UploadSystemdServiceUnitsRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_debugd_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UploadSystemdServiceUnitsResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_debugd_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_debugd_proto_goTypes,
		DependencyIndexes: file_debugd_proto_depIdxs,
		EnumInfos:         file_debugd_proto_enumTypes,
		MessageInfos:      file_debugd_proto_msgTypes,
	}.Build()
	File_debugd_proto = out.File
	file_debugd_proto_rawDesc = nil
	file_debugd_proto_goTypes = nil
	file_debugd_proto_depIdxs = nil
}
