// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.8
// source: debugd.proto

package service

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// DebugdClient is the client API for Debugd service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DebugdClient interface {
	SetInfo(ctx context.Context, in *SetInfoRequest, opts ...grpc.CallOption) (*SetInfoResponse, error)
	GetInfo(ctx context.Context, in *GetInfoRequest, opts ...grpc.CallOption) (*GetInfoResponse, error)
	UploadBootstrapper(ctx context.Context, opts ...grpc.CallOption) (Debugd_UploadBootstrapperClient, error)
	DownloadBootstrapper(ctx context.Context, in *DownloadBootstrapperRequest, opts ...grpc.CallOption) (Debugd_DownloadBootstrapperClient, error)
	UploadSystemServiceUnits(ctx context.Context, in *UploadSystemdServiceUnitsRequest, opts ...grpc.CallOption) (*UploadSystemdServiceUnitsResponse, error)
}

type debugdClient struct {
	cc grpc.ClientConnInterface
}

func NewDebugdClient(cc grpc.ClientConnInterface) DebugdClient {
	return &debugdClient{cc}
}

func (c *debugdClient) SetInfo(ctx context.Context, in *SetInfoRequest, opts ...grpc.CallOption) (*SetInfoResponse, error) {
	out := new(SetInfoResponse)
	err := c.cc.Invoke(ctx, "/debugd.Debugd/SetInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *debugdClient) GetInfo(ctx context.Context, in *GetInfoRequest, opts ...grpc.CallOption) (*GetInfoResponse, error) {
	out := new(GetInfoResponse)
	err := c.cc.Invoke(ctx, "/debugd.Debugd/GetInfo", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *debugdClient) UploadBootstrapper(ctx context.Context, opts ...grpc.CallOption) (Debugd_UploadBootstrapperClient, error) {
	stream, err := c.cc.NewStream(ctx, &Debugd_ServiceDesc.Streams[0], "/debugd.Debugd/UploadBootstrapper", opts...)
	if err != nil {
		return nil, err
	}
	x := &debugdUploadBootstrapperClient{stream}
	return x, nil
}

type Debugd_UploadBootstrapperClient interface {
	Send(*Chunk) error
	CloseAndRecv() (*UploadBootstrapperResponse, error)
	grpc.ClientStream
}

type debugdUploadBootstrapperClient struct {
	grpc.ClientStream
}

func (x *debugdUploadBootstrapperClient) Send(m *Chunk) error {
	return x.ClientStream.SendMsg(m)
}

func (x *debugdUploadBootstrapperClient) CloseAndRecv() (*UploadBootstrapperResponse, error) {
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	m := new(UploadBootstrapperResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *debugdClient) DownloadBootstrapper(ctx context.Context, in *DownloadBootstrapperRequest, opts ...grpc.CallOption) (Debugd_DownloadBootstrapperClient, error) {
	stream, err := c.cc.NewStream(ctx, &Debugd_ServiceDesc.Streams[1], "/debugd.Debugd/DownloadBootstrapper", opts...)
	if err != nil {
		return nil, err
	}
	x := &debugdDownloadBootstrapperClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Debugd_DownloadBootstrapperClient interface {
	Recv() (*Chunk, error)
	grpc.ClientStream
}

type debugdDownloadBootstrapperClient struct {
	grpc.ClientStream
}

func (x *debugdDownloadBootstrapperClient) Recv() (*Chunk, error) {
	m := new(Chunk)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *debugdClient) UploadSystemServiceUnits(ctx context.Context, in *UploadSystemdServiceUnitsRequest, opts ...grpc.CallOption) (*UploadSystemdServiceUnitsResponse, error) {
	out := new(UploadSystemdServiceUnitsResponse)
	err := c.cc.Invoke(ctx, "/debugd.Debugd/UploadSystemServiceUnits", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DebugdServer is the server API for Debugd service.
// All implementations must embed UnimplementedDebugdServer
// for forward compatibility
type DebugdServer interface {
	SetInfo(context.Context, *SetInfoRequest) (*SetInfoResponse, error)
	GetInfo(context.Context, *GetInfoRequest) (*GetInfoResponse, error)
	UploadBootstrapper(Debugd_UploadBootstrapperServer) error
	DownloadBootstrapper(*DownloadBootstrapperRequest, Debugd_DownloadBootstrapperServer) error
	UploadSystemServiceUnits(context.Context, *UploadSystemdServiceUnitsRequest) (*UploadSystemdServiceUnitsResponse, error)
	mustEmbedUnimplementedDebugdServer()
}

// UnimplementedDebugdServer must be embedded to have forward compatible implementations.
type UnimplementedDebugdServer struct {
}

func (UnimplementedDebugdServer) SetInfo(context.Context, *SetInfoRequest) (*SetInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetInfo not implemented")
}
func (UnimplementedDebugdServer) GetInfo(context.Context, *GetInfoRequest) (*GetInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetInfo not implemented")
}
func (UnimplementedDebugdServer) UploadBootstrapper(Debugd_UploadBootstrapperServer) error {
	return status.Errorf(codes.Unimplemented, "method UploadBootstrapper not implemented")
}
func (UnimplementedDebugdServer) DownloadBootstrapper(*DownloadBootstrapperRequest, Debugd_DownloadBootstrapperServer) error {
	return status.Errorf(codes.Unimplemented, "method DownloadBootstrapper not implemented")
}
func (UnimplementedDebugdServer) UploadSystemServiceUnits(context.Context, *UploadSystemdServiceUnitsRequest) (*UploadSystemdServiceUnitsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UploadSystemServiceUnits not implemented")
}
func (UnimplementedDebugdServer) mustEmbedUnimplementedDebugdServer() {}

// UnsafeDebugdServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DebugdServer will
// result in compilation errors.
type UnsafeDebugdServer interface {
	mustEmbedUnimplementedDebugdServer()
}

func RegisterDebugdServer(s grpc.ServiceRegistrar, srv DebugdServer) {
	s.RegisterService(&Debugd_ServiceDesc, srv)
}

func _Debugd_SetInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DebugdServer).SetInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/debugd.Debugd/SetInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DebugdServer).SetInfo(ctx, req.(*SetInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Debugd_GetInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DebugdServer).GetInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/debugd.Debugd/GetInfo",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DebugdServer).GetInfo(ctx, req.(*GetInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Debugd_UploadBootstrapper_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(DebugdServer).UploadBootstrapper(&debugdUploadBootstrapperServer{stream})
}

type Debugd_UploadBootstrapperServer interface {
	SendAndClose(*UploadBootstrapperResponse) error
	Recv() (*Chunk, error)
	grpc.ServerStream
}

type debugdUploadBootstrapperServer struct {
	grpc.ServerStream
}

func (x *debugdUploadBootstrapperServer) SendAndClose(m *UploadBootstrapperResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *debugdUploadBootstrapperServer) Recv() (*Chunk, error) {
	m := new(Chunk)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Debugd_DownloadBootstrapper_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(DownloadBootstrapperRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(DebugdServer).DownloadBootstrapper(m, &debugdDownloadBootstrapperServer{stream})
}

type Debugd_DownloadBootstrapperServer interface {
	Send(*Chunk) error
	grpc.ServerStream
}

type debugdDownloadBootstrapperServer struct {
	grpc.ServerStream
}

func (x *debugdDownloadBootstrapperServer) Send(m *Chunk) error {
	return x.ServerStream.SendMsg(m)
}

func _Debugd_UploadSystemServiceUnits_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UploadSystemdServiceUnitsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DebugdServer).UploadSystemServiceUnits(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/debugd.Debugd/UploadSystemServiceUnits",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DebugdServer).UploadSystemServiceUnits(ctx, req.(*UploadSystemdServiceUnitsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Debugd_ServiceDesc is the grpc.ServiceDesc for Debugd service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Debugd_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "debugd.Debugd",
	HandlerType: (*DebugdServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SetInfo",
			Handler:    _Debugd_SetInfo_Handler,
		},
		{
			MethodName: "GetInfo",
			Handler:    _Debugd_GetInfo_Handler,
		},
		{
			MethodName: "UploadSystemServiceUnits",
			Handler:    _Debugd_UploadSystemServiceUnits_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "UploadBootstrapper",
			Handler:       _Debugd_UploadBootstrapper_Handler,
			ClientStreams: true,
		},
		{
			StreamName:    "DownloadBootstrapper",
			Handler:       _Debugd_DownloadBootstrapper_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "debugd.proto",
}
