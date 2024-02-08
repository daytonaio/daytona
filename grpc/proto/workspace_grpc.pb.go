// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.12.4
// source: grpc/proto/workspace.proto

package proto

import (
	context "context"
	types "github.com/daytonaio/daytona/grpc/proto/types"
	empty "github.com/golang/protobuf/ptypes/empty"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	WorkspaceService_Create_FullMethodName = "/WorkspaceService/Create"
	WorkspaceService_Info_FullMethodName   = "/WorkspaceService/Info"
	WorkspaceService_List_FullMethodName   = "/WorkspaceService/List"
	WorkspaceService_Start_FullMethodName  = "/WorkspaceService/Start"
	WorkspaceService_Stop_FullMethodName   = "/WorkspaceService/Stop"
	WorkspaceService_Remove_FullMethodName = "/WorkspaceService/Remove"
)

// WorkspaceServiceClient is the client API for WorkspaceService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type WorkspaceServiceClient interface {
	Create(ctx context.Context, in *CreateWorkspaceRequest, opts ...grpc.CallOption) (WorkspaceService_CreateClient, error)
	Info(ctx context.Context, in *WorkspaceInfoRequest, opts ...grpc.CallOption) (*types.WorkspaceInfo, error)
	List(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*WorkspaceListResponse, error)
	Start(ctx context.Context, in *WorkspaceStartRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	Stop(ctx context.Context, in *WorkspaceStopRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	Remove(ctx context.Context, in *WorkspaceRemoveRequest, opts ...grpc.CallOption) (*empty.Empty, error)
}

type workspaceServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewWorkspaceServiceClient(cc grpc.ClientConnInterface) WorkspaceServiceClient {
	return &workspaceServiceClient{cc}
}

func (c *workspaceServiceClient) Create(ctx context.Context, in *CreateWorkspaceRequest, opts ...grpc.CallOption) (WorkspaceService_CreateClient, error) {
	stream, err := c.cc.NewStream(ctx, &WorkspaceService_ServiceDesc.Streams[0], WorkspaceService_Create_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &workspaceServiceCreateClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type WorkspaceService_CreateClient interface {
	Recv() (*CreateWorkspaceResponse, error)
	grpc.ClientStream
}

type workspaceServiceCreateClient struct {
	grpc.ClientStream
}

func (x *workspaceServiceCreateClient) Recv() (*CreateWorkspaceResponse, error) {
	m := new(CreateWorkspaceResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *workspaceServiceClient) Info(ctx context.Context, in *WorkspaceInfoRequest, opts ...grpc.CallOption) (*types.WorkspaceInfo, error) {
	out := new(types.WorkspaceInfo)
	err := c.cc.Invoke(ctx, WorkspaceService_Info_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workspaceServiceClient) List(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*WorkspaceListResponse, error) {
	out := new(WorkspaceListResponse)
	err := c.cc.Invoke(ctx, WorkspaceService_List_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workspaceServiceClient) Start(ctx context.Context, in *WorkspaceStartRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, WorkspaceService_Start_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workspaceServiceClient) Stop(ctx context.Context, in *WorkspaceStopRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, WorkspaceService_Stop_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *workspaceServiceClient) Remove(ctx context.Context, in *WorkspaceRemoveRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, WorkspaceService_Remove_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// WorkspaceServiceServer is the server API for WorkspaceService service.
// All implementations should embed UnimplementedWorkspaceServiceServer
// for forward compatibility
type WorkspaceServiceServer interface {
	Create(*CreateWorkspaceRequest, WorkspaceService_CreateServer) error
	Info(context.Context, *WorkspaceInfoRequest) (*types.WorkspaceInfo, error)
	List(context.Context, *empty.Empty) (*WorkspaceListResponse, error)
	Start(context.Context, *WorkspaceStartRequest) (*empty.Empty, error)
	Stop(context.Context, *WorkspaceStopRequest) (*empty.Empty, error)
	Remove(context.Context, *WorkspaceRemoveRequest) (*empty.Empty, error)
}

// UnimplementedWorkspaceServiceServer should be embedded to have forward compatible implementations.
type UnimplementedWorkspaceServiceServer struct {
}

func (UnimplementedWorkspaceServiceServer) Create(*CreateWorkspaceRequest, WorkspaceService_CreateServer) error {
	return status.Errorf(codes.Unimplemented, "method Create not implemented")
}
func (UnimplementedWorkspaceServiceServer) Info(context.Context, *WorkspaceInfoRequest) (*types.WorkspaceInfo, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Info not implemented")
}
func (UnimplementedWorkspaceServiceServer) List(context.Context, *empty.Empty) (*WorkspaceListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
func (UnimplementedWorkspaceServiceServer) Start(context.Context, *WorkspaceStartRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Start not implemented")
}
func (UnimplementedWorkspaceServiceServer) Stop(context.Context, *WorkspaceStopRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Stop not implemented")
}
func (UnimplementedWorkspaceServiceServer) Remove(context.Context, *WorkspaceRemoveRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Remove not implemented")
}

// UnsafeWorkspaceServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to WorkspaceServiceServer will
// result in compilation errors.
type UnsafeWorkspaceServiceServer interface {
	mustEmbedUnimplementedWorkspaceServiceServer()
}

func RegisterWorkspaceServiceServer(s grpc.ServiceRegistrar, srv WorkspaceServiceServer) {
	s.RegisterService(&WorkspaceService_ServiceDesc, srv)
}

func _WorkspaceService_Create_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(CreateWorkspaceRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(WorkspaceServiceServer).Create(m, &workspaceServiceCreateServer{stream})
}

type WorkspaceService_CreateServer interface {
	Send(*CreateWorkspaceResponse) error
	grpc.ServerStream
}

type workspaceServiceCreateServer struct {
	grpc.ServerStream
}

func (x *workspaceServiceCreateServer) Send(m *CreateWorkspaceResponse) error {
	return x.ServerStream.SendMsg(m)
}

func _WorkspaceService_Info_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WorkspaceInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkspaceServiceServer).Info(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorkspaceService_Info_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkspaceServiceServer).Info(ctx, req.(*WorkspaceInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkspaceService_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkspaceServiceServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorkspaceService_List_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkspaceServiceServer).List(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkspaceService_Start_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WorkspaceStartRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkspaceServiceServer).Start(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorkspaceService_Start_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkspaceServiceServer).Start(ctx, req.(*WorkspaceStartRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkspaceService_Stop_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WorkspaceStopRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkspaceServiceServer).Stop(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorkspaceService_Stop_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkspaceServiceServer).Stop(ctx, req.(*WorkspaceStopRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _WorkspaceService_Remove_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(WorkspaceRemoveRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(WorkspaceServiceServer).Remove(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: WorkspaceService_Remove_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(WorkspaceServiceServer).Remove(ctx, req.(*WorkspaceRemoveRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// WorkspaceService_ServiceDesc is the grpc.ServiceDesc for WorkspaceService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var WorkspaceService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "WorkspaceService",
	HandlerType: (*WorkspaceServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Info",
			Handler:    _WorkspaceService_Info_Handler,
		},
		{
			MethodName: "List",
			Handler:    _WorkspaceService_List_Handler,
		},
		{
			MethodName: "Start",
			Handler:    _WorkspaceService_Start_Handler,
		},
		{
			MethodName: "Stop",
			Handler:    _WorkspaceService_Stop_Handler,
		},
		{
			MethodName: "Remove",
			Handler:    _WorkspaceService_Remove_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Create",
			Handler:       _WorkspaceService_Create_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "grpc/proto/workspace.proto",
}
