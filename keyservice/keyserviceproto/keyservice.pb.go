// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.5
// 	protoc        v5.29.1
// source: keyservice/keyserviceproto/keyservice.proto

package keyserviceproto

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type GetDataKeyRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	DataKeyId     string                 `protobuf:"bytes,1,opt,name=data_key_id,json=dataKeyId,proto3" json:"data_key_id,omitempty"`
	Length        uint32                 `protobuf:"varint,2,opt,name=length,proto3" json:"length,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetDataKeyRequest) Reset() {
	*x = GetDataKeyRequest{}
	mi := &file_keyservice_keyserviceproto_keyservice_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetDataKeyRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDataKeyRequest) ProtoMessage() {}

func (x *GetDataKeyRequest) ProtoReflect() protoreflect.Message {
	mi := &file_keyservice_keyserviceproto_keyservice_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetDataKeyRequest.ProtoReflect.Descriptor instead.
func (*GetDataKeyRequest) Descriptor() ([]byte, []int) {
	return file_keyservice_keyserviceproto_keyservice_proto_rawDescGZIP(), []int{0}
}

func (x *GetDataKeyRequest) GetDataKeyId() string {
	if x != nil {
		return x.DataKeyId
	}
	return ""
}

func (x *GetDataKeyRequest) GetLength() uint32 {
	if x != nil {
		return x.Length
	}
	return 0
}

type GetDataKeyResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	DataKey       []byte                 `protobuf:"bytes,1,opt,name=data_key,json=dataKey,proto3" json:"data_key,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetDataKeyResponse) Reset() {
	*x = GetDataKeyResponse{}
	mi := &file_keyservice_keyserviceproto_keyservice_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetDataKeyResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetDataKeyResponse) ProtoMessage() {}

func (x *GetDataKeyResponse) ProtoReflect() protoreflect.Message {
	mi := &file_keyservice_keyserviceproto_keyservice_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetDataKeyResponse.ProtoReflect.Descriptor instead.
func (*GetDataKeyResponse) Descriptor() ([]byte, []int) {
	return file_keyservice_keyserviceproto_keyservice_proto_rawDescGZIP(), []int{1}
}

func (x *GetDataKeyResponse) GetDataKey() []byte {
	if x != nil {
		return x.DataKey
	}
	return nil
}

var File_keyservice_keyserviceproto_keyservice_proto protoreflect.FileDescriptor

var file_keyservice_keyserviceproto_keyservice_proto_rawDesc = string([]byte{
	0x0a, 0x2b, 0x6b, 0x65, 0x79, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x6b, 0x65, 0x79,
	0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6b, 0x65, 0x79,
	0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x03, 0x6b,
	0x6d, 0x73, 0x22, 0x4b, 0x0a, 0x11, 0x47, 0x65, 0x74, 0x44, 0x61, 0x74, 0x61, 0x4b, 0x65, 0x79,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1e, 0x0a, 0x0b, 0x64, 0x61, 0x74, 0x61, 0x5f,
	0x6b, 0x65, 0x79, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x64, 0x61,
	0x74, 0x61, 0x4b, 0x65, 0x79, 0x49, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x6c, 0x65, 0x6e, 0x67, 0x74,
	0x68, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x06, 0x6c, 0x65, 0x6e, 0x67, 0x74, 0x68, 0x22,
	0x2f, 0x0a, 0x12, 0x47, 0x65, 0x74, 0x44, 0x61, 0x74, 0x61, 0x4b, 0x65, 0x79, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x19, 0x0a, 0x08, 0x64, 0x61, 0x74, 0x61, 0x5f, 0x6b, 0x65,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x07, 0x64, 0x61, 0x74, 0x61, 0x4b, 0x65, 0x79,
	0x32, 0x44, 0x0a, 0x03, 0x41, 0x50, 0x49, 0x12, 0x3d, 0x0a, 0x0a, 0x47, 0x65, 0x74, 0x44, 0x61,
	0x74, 0x61, 0x4b, 0x65, 0x79, 0x12, 0x16, 0x2e, 0x6b, 0x6d, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x44,
	0x61, 0x74, 0x61, 0x4b, 0x65, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x17, 0x2e,
	0x6b, 0x6d, 0x73, 0x2e, 0x47, 0x65, 0x74, 0x44, 0x61, 0x74, 0x61, 0x4b, 0x65, 0x79, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x44, 0x5a, 0x42, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x65, 0x64, 0x67, 0x65, 0x6c, 0x65, 0x73, 0x73, 0x73, 0x79, 0x73,
	0x2f, 0x63, 0x6f, 0x6e, 0x73, 0x74, 0x65, 0x6c, 0x6c, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x76,
	0x32, 0x2f, 0x6b, 0x65, 0x79, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x6b, 0x65, 0x79,
	0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_keyservice_keyserviceproto_keyservice_proto_rawDescOnce sync.Once
	file_keyservice_keyserviceproto_keyservice_proto_rawDescData []byte
)

func file_keyservice_keyserviceproto_keyservice_proto_rawDescGZIP() []byte {
	file_keyservice_keyserviceproto_keyservice_proto_rawDescOnce.Do(func() {
		file_keyservice_keyserviceproto_keyservice_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_keyservice_keyserviceproto_keyservice_proto_rawDesc), len(file_keyservice_keyserviceproto_keyservice_proto_rawDesc)))
	})
	return file_keyservice_keyserviceproto_keyservice_proto_rawDescData
}

var file_keyservice_keyserviceproto_keyservice_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_keyservice_keyserviceproto_keyservice_proto_goTypes = []any{
	(*GetDataKeyRequest)(nil),  // 0: kms.GetDataKeyRequest
	(*GetDataKeyResponse)(nil), // 1: kms.GetDataKeyResponse
}
var file_keyservice_keyserviceproto_keyservice_proto_depIdxs = []int32{
	0, // 0: kms.API.GetDataKey:input_type -> kms.GetDataKeyRequest
	1, // 1: kms.API.GetDataKey:output_type -> kms.GetDataKeyResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_keyservice_keyserviceproto_keyservice_proto_init() }
func file_keyservice_keyserviceproto_keyservice_proto_init() {
	if File_keyservice_keyserviceproto_keyservice_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_keyservice_keyserviceproto_keyservice_proto_rawDesc), len(file_keyservice_keyserviceproto_keyservice_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_keyservice_keyserviceproto_keyservice_proto_goTypes,
		DependencyIndexes: file_keyservice_keyserviceproto_keyservice_proto_depIdxs,
		MessageInfos:      file_keyservice_keyserviceproto_keyservice_proto_msgTypes,
	}.Build()
	File_keyservice_keyserviceproto_keyservice_proto = out.File
	file_keyservice_keyserviceproto_keyservice_proto_goTypes = nil
	file_keyservice_keyserviceproto_keyservice_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// APIClient is the client API for API service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type APIClient interface {
	GetDataKey(ctx context.Context, in *GetDataKeyRequest, opts ...grpc.CallOption) (*GetDataKeyResponse, error)
}

type aPIClient struct {
	cc grpc.ClientConnInterface
}

func NewAPIClient(cc grpc.ClientConnInterface) APIClient {
	return &aPIClient{cc}
}

func (c *aPIClient) GetDataKey(ctx context.Context, in *GetDataKeyRequest, opts ...grpc.CallOption) (*GetDataKeyResponse, error) {
	out := new(GetDataKeyResponse)
	err := c.cc.Invoke(ctx, "/kms.API/GetDataKey", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// APIServer is the server API for API service.
type APIServer interface {
	GetDataKey(context.Context, *GetDataKeyRequest) (*GetDataKeyResponse, error)
}

// UnimplementedAPIServer can be embedded to have forward compatible implementations.
type UnimplementedAPIServer struct {
}

func (*UnimplementedAPIServer) GetDataKey(context.Context, *GetDataKeyRequest) (*GetDataKeyResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetDataKey not implemented")
}

func RegisterAPIServer(s *grpc.Server, srv APIServer) {
	s.RegisterService(&_API_serviceDesc, srv)
}

func _API_GetDataKey_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetDataKeyRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(APIServer).GetDataKey(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/kms.API/GetDataKey",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(APIServer).GetDataKey(ctx, req.(*GetDataKeyRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _API_serviceDesc = grpc.ServiceDesc{
	ServiceName: "kms.API",
	HandlerType: (*APIServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetDataKey",
			Handler:    _API_GetDataKey_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "keyservice/keyserviceproto/keyservice.proto",
}
