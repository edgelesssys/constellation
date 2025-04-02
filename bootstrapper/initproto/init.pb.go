// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.1
// source: bootstrapper/initproto/init.proto

package initproto

import (
	context "context"
	components "github.com/edgelesssys/constellation/v2/internal/versions/components"
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

type InitRequest struct {
	state                protoimpl.MessageState  `protogen:"open.v1"`
	KmsUri               string                  `protobuf:"bytes,1,opt,name=kms_uri,json=kmsUri,proto3" json:"kms_uri,omitempty"`
	StorageUri           string                  `protobuf:"bytes,2,opt,name=storage_uri,json=storageUri,proto3" json:"storage_uri,omitempty"`
	MeasurementSalt      []byte                  `protobuf:"bytes,3,opt,name=measurement_salt,json=measurementSalt,proto3" json:"measurement_salt,omitempty"`
	KubernetesVersion    string                  `protobuf:"bytes,5,opt,name=kubernetes_version,json=kubernetesVersion,proto3" json:"kubernetes_version,omitempty"`
	ConformanceMode      bool                    `protobuf:"varint,6,opt,name=conformance_mode,json=conformanceMode,proto3" json:"conformance_mode,omitempty"`
	KubernetesComponents []*components.Component `protobuf:"bytes,7,rep,name=kubernetes_components,json=kubernetesComponents,proto3" json:"kubernetes_components,omitempty"`
	InitSecret           []byte                  `protobuf:"bytes,8,opt,name=init_secret,json=initSecret,proto3" json:"init_secret,omitempty"`
	ClusterName          string                  `protobuf:"bytes,9,opt,name=cluster_name,json=clusterName,proto3" json:"cluster_name,omitempty"`
	ApiserverCertSans    []string                `protobuf:"bytes,10,rep,name=apiserver_cert_sans,json=apiserverCertSans,proto3" json:"apiserver_cert_sans,omitempty"`
	ServiceCidr          string                  `protobuf:"bytes,11,opt,name=service_cidr,json=serviceCidr,proto3" json:"service_cidr,omitempty"`
	unknownFields        protoimpl.UnknownFields
	sizeCache            protoimpl.SizeCache
}

func (x *InitRequest) Reset() {
	*x = InitRequest{}
	mi := &file_bootstrapper_initproto_init_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *InitRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InitRequest) ProtoMessage() {}

func (x *InitRequest) ProtoReflect() protoreflect.Message {
	mi := &file_bootstrapper_initproto_init_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InitRequest.ProtoReflect.Descriptor instead.
func (*InitRequest) Descriptor() ([]byte, []int) {
	return file_bootstrapper_initproto_init_proto_rawDescGZIP(), []int{0}
}

func (x *InitRequest) GetKmsUri() string {
	if x != nil {
		return x.KmsUri
	}
	return ""
}

func (x *InitRequest) GetStorageUri() string {
	if x != nil {
		return x.StorageUri
	}
	return ""
}

func (x *InitRequest) GetMeasurementSalt() []byte {
	if x != nil {
		return x.MeasurementSalt
	}
	return nil
}

func (x *InitRequest) GetKubernetesVersion() string {
	if x != nil {
		return x.KubernetesVersion
	}
	return ""
}

func (x *InitRequest) GetConformanceMode() bool {
	if x != nil {
		return x.ConformanceMode
	}
	return false
}

func (x *InitRequest) GetKubernetesComponents() []*components.Component {
	if x != nil {
		return x.KubernetesComponents
	}
	return nil
}

func (x *InitRequest) GetInitSecret() []byte {
	if x != nil {
		return x.InitSecret
	}
	return nil
}

func (x *InitRequest) GetClusterName() string {
	if x != nil {
		return x.ClusterName
	}
	return ""
}

func (x *InitRequest) GetApiserverCertSans() []string {
	if x != nil {
		return x.ApiserverCertSans
	}
	return nil
}

func (x *InitRequest) GetServiceCidr() string {
	if x != nil {
		return x.ServiceCidr
	}
	return ""
}

type InitResponse struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Types that are valid to be assigned to Kind:
	//
	//	*InitResponse_InitSuccess
	//	*InitResponse_InitFailure
	//	*InitResponse_Log
	Kind          isInitResponse_Kind `protobuf_oneof:"kind"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *InitResponse) Reset() {
	*x = InitResponse{}
	mi := &file_bootstrapper_initproto_init_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *InitResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InitResponse) ProtoMessage() {}

func (x *InitResponse) ProtoReflect() protoreflect.Message {
	mi := &file_bootstrapper_initproto_init_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InitResponse.ProtoReflect.Descriptor instead.
func (*InitResponse) Descriptor() ([]byte, []int) {
	return file_bootstrapper_initproto_init_proto_rawDescGZIP(), []int{1}
}

func (x *InitResponse) GetKind() isInitResponse_Kind {
	if x != nil {
		return x.Kind
	}
	return nil
}

func (x *InitResponse) GetInitSuccess() *InitSuccessResponse {
	if x != nil {
		if x, ok := x.Kind.(*InitResponse_InitSuccess); ok {
			return x.InitSuccess
		}
	}
	return nil
}

func (x *InitResponse) GetInitFailure() *InitFailureResponse {
	if x != nil {
		if x, ok := x.Kind.(*InitResponse_InitFailure); ok {
			return x.InitFailure
		}
	}
	return nil
}

func (x *InitResponse) GetLog() *LogResponseType {
	if x != nil {
		if x, ok := x.Kind.(*InitResponse_Log); ok {
			return x.Log
		}
	}
	return nil
}

type isInitResponse_Kind interface {
	isInitResponse_Kind()
}

type InitResponse_InitSuccess struct {
	InitSuccess *InitSuccessResponse `protobuf:"bytes,1,opt,name=init_success,json=initSuccess,proto3,oneof"`
}

type InitResponse_InitFailure struct {
	InitFailure *InitFailureResponse `protobuf:"bytes,2,opt,name=init_failure,json=initFailure,proto3,oneof"`
}

type InitResponse_Log struct {
	Log *LogResponseType `protobuf:"bytes,3,opt,name=log,proto3,oneof"`
}

func (*InitResponse_InitSuccess) isInitResponse_Kind() {}

func (*InitResponse_InitFailure) isInitResponse_Kind() {}

func (*InitResponse_Log) isInitResponse_Kind() {}

type InitSuccessResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Kubeconfig    []byte                 `protobuf:"bytes,1,opt,name=kubeconfig,proto3" json:"kubeconfig,omitempty"`
	OwnerId       []byte                 `protobuf:"bytes,2,opt,name=owner_id,json=ownerId,proto3" json:"owner_id,omitempty"`
	ClusterId     []byte                 `protobuf:"bytes,3,opt,name=cluster_id,json=clusterId,proto3" json:"cluster_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *InitSuccessResponse) Reset() {
	*x = InitSuccessResponse{}
	mi := &file_bootstrapper_initproto_init_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *InitSuccessResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InitSuccessResponse) ProtoMessage() {}

func (x *InitSuccessResponse) ProtoReflect() protoreflect.Message {
	mi := &file_bootstrapper_initproto_init_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InitSuccessResponse.ProtoReflect.Descriptor instead.
func (*InitSuccessResponse) Descriptor() ([]byte, []int) {
	return file_bootstrapper_initproto_init_proto_rawDescGZIP(), []int{2}
}

func (x *InitSuccessResponse) GetKubeconfig() []byte {
	if x != nil {
		return x.Kubeconfig
	}
	return nil
}

func (x *InitSuccessResponse) GetOwnerId() []byte {
	if x != nil {
		return x.OwnerId
	}
	return nil
}

func (x *InitSuccessResponse) GetClusterId() []byte {
	if x != nil {
		return x.ClusterId
	}
	return nil
}

type InitFailureResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Error         string                 `protobuf:"bytes,1,opt,name=error,proto3" json:"error,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *InitFailureResponse) Reset() {
	*x = InitFailureResponse{}
	mi := &file_bootstrapper_initproto_init_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *InitFailureResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InitFailureResponse) ProtoMessage() {}

func (x *InitFailureResponse) ProtoReflect() protoreflect.Message {
	mi := &file_bootstrapper_initproto_init_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InitFailureResponse.ProtoReflect.Descriptor instead.
func (*InitFailureResponse) Descriptor() ([]byte, []int) {
	return file_bootstrapper_initproto_init_proto_rawDescGZIP(), []int{3}
}

func (x *InitFailureResponse) GetError() string {
	if x != nil {
		return x.Error
	}
	return ""
}

type LogResponseType struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Log           []byte                 `protobuf:"bytes,1,opt,name=log,proto3" json:"log,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *LogResponseType) Reset() {
	*x = LogResponseType{}
	mi := &file_bootstrapper_initproto_init_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *LogResponseType) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LogResponseType) ProtoMessage() {}

func (x *LogResponseType) ProtoReflect() protoreflect.Message {
	mi := &file_bootstrapper_initproto_init_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LogResponseType.ProtoReflect.Descriptor instead.
func (*LogResponseType) Descriptor() ([]byte, []int) {
	return file_bootstrapper_initproto_init_proto_rawDescGZIP(), []int{4}
}

func (x *LogResponseType) GetLog() []byte {
	if x != nil {
		return x.Log
	}
	return nil
}

type KubernetesComponent struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Url           string                 `protobuf:"bytes,1,opt,name=url,proto3" json:"url,omitempty"`
	Hash          string                 `protobuf:"bytes,2,opt,name=hash,proto3" json:"hash,omitempty"`
	InstallPath   string                 `protobuf:"bytes,3,opt,name=install_path,json=installPath,proto3" json:"install_path,omitempty"`
	Extract       bool                   `protobuf:"varint,4,opt,name=extract,proto3" json:"extract,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *KubernetesComponent) Reset() {
	*x = KubernetesComponent{}
	mi := &file_bootstrapper_initproto_init_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *KubernetesComponent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*KubernetesComponent) ProtoMessage() {}

func (x *KubernetesComponent) ProtoReflect() protoreflect.Message {
	mi := &file_bootstrapper_initproto_init_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use KubernetesComponent.ProtoReflect.Descriptor instead.
func (*KubernetesComponent) Descriptor() ([]byte, []int) {
	return file_bootstrapper_initproto_init_proto_rawDescGZIP(), []int{5}
}

func (x *KubernetesComponent) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *KubernetesComponent) GetHash() string {
	if x != nil {
		return x.Hash
	}
	return ""
}

func (x *KubernetesComponent) GetInstallPath() string {
	if x != nil {
		return x.InstallPath
	}
	return ""
}

func (x *KubernetesComponent) GetExtract() bool {
	if x != nil {
		return x.Extract
	}
	return false
}

var File_bootstrapper_initproto_init_proto protoreflect.FileDescriptor

const file_bootstrapper_initproto_init_proto_rawDesc = "" +
	"\n" +
	"!bootstrapper/initproto/init.proto\x12\x04init\x1a-internal/versions/components/components.proto\"\xd0\x03\n" +
	"\vInitRequest\x12\x17\n" +
	"\akms_uri\x18\x01 \x01(\tR\x06kmsUri\x12\x1f\n" +
	"\vstorage_uri\x18\x02 \x01(\tR\n" +
	"storageUri\x12)\n" +
	"\x10measurement_salt\x18\x03 \x01(\fR\x0fmeasurementSalt\x12-\n" +
	"\x12kubernetes_version\x18\x05 \x01(\tR\x11kubernetesVersion\x12)\n" +
	"\x10conformance_mode\x18\x06 \x01(\bR\x0fconformanceMode\x12J\n" +
	"\x15kubernetes_components\x18\a \x03(\v2\x15.components.ComponentR\x14kubernetesComponents\x12\x1f\n" +
	"\vinit_secret\x18\b \x01(\fR\n" +
	"initSecret\x12!\n" +
	"\fcluster_name\x18\t \x01(\tR\vclusterName\x12.\n" +
	"\x13apiserver_cert_sans\x18\n" +
	" \x03(\tR\x11apiserverCertSans\x12!\n" +
	"\fservice_cidr\x18\v \x01(\tR\vserviceCidrJ\x04\b\x04\x10\x05R\x19cloud_service_account_uri\"\xc1\x01\n" +
	"\fInitResponse\x12>\n" +
	"\finit_success\x18\x01 \x01(\v2\x19.init.InitSuccessResponseH\x00R\vinitSuccess\x12>\n" +
	"\finit_failure\x18\x02 \x01(\v2\x19.init.InitFailureResponseH\x00R\vinitFailure\x12)\n" +
	"\x03log\x18\x03 \x01(\v2\x15.init.LogResponseTypeH\x00R\x03logB\x06\n" +
	"\x04kind\"o\n" +
	"\x13InitSuccessResponse\x12\x1e\n" +
	"\n" +
	"kubeconfig\x18\x01 \x01(\fR\n" +
	"kubeconfig\x12\x19\n" +
	"\bowner_id\x18\x02 \x01(\fR\aownerId\x12\x1d\n" +
	"\n" +
	"cluster_id\x18\x03 \x01(\fR\tclusterId\"+\n" +
	"\x13InitFailureResponse\x12\x14\n" +
	"\x05error\x18\x01 \x01(\tR\x05error\"#\n" +
	"\x0fLogResponseType\x12\x10\n" +
	"\x03log\x18\x01 \x01(\fR\x03log\"x\n" +
	"\x13KubernetesComponent\x12\x10\n" +
	"\x03url\x18\x01 \x01(\tR\x03url\x12\x12\n" +
	"\x04hash\x18\x02 \x01(\tR\x04hash\x12!\n" +
	"\finstall_path\x18\x03 \x01(\tR\vinstallPath\x12\x18\n" +
	"\aextract\x18\x04 \x01(\bR\aextract26\n" +
	"\x03API\x12/\n" +
	"\x04Init\x12\x11.init.InitRequest\x1a\x12.init.InitResponse0\x01B@Z>github.com/edgelesssys/constellation/v2/bootstrapper/initprotob\x06proto3"

var (
	file_bootstrapper_initproto_init_proto_rawDescOnce sync.Once
	file_bootstrapper_initproto_init_proto_rawDescData []byte
)

func file_bootstrapper_initproto_init_proto_rawDescGZIP() []byte {
	file_bootstrapper_initproto_init_proto_rawDescOnce.Do(func() {
		file_bootstrapper_initproto_init_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_bootstrapper_initproto_init_proto_rawDesc), len(file_bootstrapper_initproto_init_proto_rawDesc)))
	})
	return file_bootstrapper_initproto_init_proto_rawDescData
}

var file_bootstrapper_initproto_init_proto_msgTypes = make([]protoimpl.MessageInfo, 6)
var file_bootstrapper_initproto_init_proto_goTypes = []any{
	(*InitRequest)(nil),          // 0: init.InitRequest
	(*InitResponse)(nil),         // 1: init.InitResponse
	(*InitSuccessResponse)(nil),  // 2: init.InitSuccessResponse
	(*InitFailureResponse)(nil),  // 3: init.InitFailureResponse
	(*LogResponseType)(nil),      // 4: init.LogResponseType
	(*KubernetesComponent)(nil),  // 5: init.KubernetesComponent
	(*components.Component)(nil), // 6: components.Component
}
var file_bootstrapper_initproto_init_proto_depIdxs = []int32{
	6, // 0: init.InitRequest.kubernetes_components:type_name -> components.Component
	2, // 1: init.InitResponse.init_success:type_name -> init.InitSuccessResponse
	3, // 2: init.InitResponse.init_failure:type_name -> init.InitFailureResponse
	4, // 3: init.InitResponse.log:type_name -> init.LogResponseType
	0, // 4: init.API.Init:input_type -> init.InitRequest
	1, // 5: init.API.Init:output_type -> init.InitResponse
	5, // [5:6] is the sub-list for method output_type
	4, // [4:5] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_bootstrapper_initproto_init_proto_init() }
func file_bootstrapper_initproto_init_proto_init() {
	if File_bootstrapper_initproto_init_proto != nil {
		return
	}
	file_bootstrapper_initproto_init_proto_msgTypes[1].OneofWrappers = []any{
		(*InitResponse_InitSuccess)(nil),
		(*InitResponse_InitFailure)(nil),
		(*InitResponse_Log)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_bootstrapper_initproto_init_proto_rawDesc), len(file_bootstrapper_initproto_init_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   6,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_bootstrapper_initproto_init_proto_goTypes,
		DependencyIndexes: file_bootstrapper_initproto_init_proto_depIdxs,
		MessageInfos:      file_bootstrapper_initproto_init_proto_msgTypes,
	}.Build()
	File_bootstrapper_initproto_init_proto = out.File
	file_bootstrapper_initproto_init_proto_goTypes = nil
	file_bootstrapper_initproto_init_proto_depIdxs = nil
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
	Init(ctx context.Context, in *InitRequest, opts ...grpc.CallOption) (API_InitClient, error)
}

type aPIClient struct {
	cc grpc.ClientConnInterface
}

func NewAPIClient(cc grpc.ClientConnInterface) APIClient {
	return &aPIClient{cc}
}

func (c *aPIClient) Init(ctx context.Context, in *InitRequest, opts ...grpc.CallOption) (API_InitClient, error) {
	stream, err := c.cc.NewStream(ctx, &_API_serviceDesc.Streams[0], "/init.API/Init", opts...)
	if err != nil {
		return nil, err
	}
	x := &aPIInitClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type API_InitClient interface {
	Recv() (*InitResponse, error)
	grpc.ClientStream
}

type aPIInitClient struct {
	grpc.ClientStream
}

func (x *aPIInitClient) Recv() (*InitResponse, error) {
	m := new(InitResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// APIServer is the server API for API service.
type APIServer interface {
	Init(*InitRequest, API_InitServer) error
}

// UnimplementedAPIServer can be embedded to have forward compatible implementations.
type UnimplementedAPIServer struct {
}

func (*UnimplementedAPIServer) Init(*InitRequest, API_InitServer) error {
	return status.Errorf(codes.Unimplemented, "method Init not implemented")
}

func RegisterAPIServer(s *grpc.Server, srv APIServer) {
	s.RegisterService(&_API_serviceDesc, srv)
}

func _API_Init_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(InitRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(APIServer).Init(m, &aPIInitServer{stream})
}

type API_InitServer interface {
	Send(*InitResponse) error
	grpc.ServerStream
}

type aPIInitServer struct {
	grpc.ServerStream
}

func (x *aPIInitServer) Send(m *InitResponse) error {
	return x.ServerStream.SendMsg(m)
}

var _API_serviceDesc = grpc.ServiceDesc{
	ServiceName: "init.API",
	HandlerType: (*APIServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Init",
			Handler:       _API_Init_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "bootstrapper/initproto/init.proto",
}
