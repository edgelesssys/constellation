// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.21.8
// source: upgrade.proto

package upgradeproto

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

// UpdateClient is the client API for Update service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UpdateClient interface {
	ExecuteUpdate(ctx context.Context, in *ExecuteUpdateRequest, opts ...grpc.CallOption) (*ExecuteUpdateResponse, error)
}

type updateClient struct {
	cc grpc.ClientConnInterface
}

func NewUpdateClient(cc grpc.ClientConnInterface) UpdateClient {
	return &updateClient{cc}
}

func (c *updateClient) ExecuteUpdate(ctx context.Context, in *ExecuteUpdateRequest, opts ...grpc.CallOption) (*ExecuteUpdateResponse, error) {
	out := new(ExecuteUpdateResponse)
	err := c.cc.Invoke(ctx, "/upgrade.Update/ExecuteUpdate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateServer is the server API for Update service.
// All implementations must embed UnimplementedUpdateServer
// for forward compatibility
type UpdateServer interface {
	ExecuteUpdate(context.Context, *ExecuteUpdateRequest) (*ExecuteUpdateResponse, error)
	mustEmbedUnimplementedUpdateServer()
}

// UnimplementedUpdateServer must be embedded to have forward compatible implementations.
type UnimplementedUpdateServer struct {
}

func (UnimplementedUpdateServer) ExecuteUpdate(context.Context, *ExecuteUpdateRequest) (*ExecuteUpdateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ExecuteUpdate not implemented")
}
func (UnimplementedUpdateServer) mustEmbedUnimplementedUpdateServer() {}

// UnsafeUpdateServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UpdateServer will
// result in compilation errors.
type UnsafeUpdateServer interface {
	mustEmbedUnimplementedUpdateServer()
}

func RegisterUpdateServer(s grpc.ServiceRegistrar, srv UpdateServer) {
	s.RegisterService(&Update_ServiceDesc, srv)
}

func _Update_ExecuteUpdate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ExecuteUpdateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UpdateServer).ExecuteUpdate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/upgrade.Update/ExecuteUpdate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UpdateServer).ExecuteUpdate(ctx, req.(*ExecuteUpdateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Update_ServiceDesc is the grpc.ServiceDesc for Update service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Update_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "upgrade.Update",
	HandlerType: (*UpdateServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ExecuteUpdate",
			Handler:    _Update_ExecuteUpdate_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "upgrade.proto",
}
