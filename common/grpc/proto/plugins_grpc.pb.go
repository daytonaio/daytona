// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v3.12.4
// source: common/grpc/proto/plugins.proto

package proto

import (
	context "context"
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
	Plugins_ListProvisionerPlugins_FullMethodName      = "/Plugins/ListProvisionerPlugins"
	Plugins_ListAgentServicePlugins_FullMethodName     = "/Plugins/ListAgentServicePlugins"
	Plugins_InstallProvisionerPlugin_FullMethodName    = "/Plugins/InstallProvisionerPlugin"
	Plugins_InstallAgentServicePlugin_FullMethodName   = "/Plugins/InstallAgentServicePlugin"
	Plugins_UninstallProvisionerPlugin_FullMethodName  = "/Plugins/UninstallProvisionerPlugin"
	Plugins_UninstallAgentServicePlugin_FullMethodName = "/Plugins/UninstallAgentServicePlugin"
)

// PluginsClient is the client API for Plugins service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PluginsClient interface {
	ListProvisionerPlugins(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*ProvisionerPluginList, error)
	ListAgentServicePlugins(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*AgentServicePluginList, error)
	InstallProvisionerPlugin(ctx context.Context, in *InstallPluginRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	InstallAgentServicePlugin(ctx context.Context, in *InstallPluginRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	UninstallProvisionerPlugin(ctx context.Context, in *UninstallPluginRequest, opts ...grpc.CallOption) (*empty.Empty, error)
	UninstallAgentServicePlugin(ctx context.Context, in *UninstallPluginRequest, opts ...grpc.CallOption) (*empty.Empty, error)
}

type pluginsClient struct {
	cc grpc.ClientConnInterface
}

func NewPluginsClient(cc grpc.ClientConnInterface) PluginsClient {
	return &pluginsClient{cc}
}

func (c *pluginsClient) ListProvisionerPlugins(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*ProvisionerPluginList, error) {
	out := new(ProvisionerPluginList)
	err := c.cc.Invoke(ctx, Plugins_ListProvisionerPlugins_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pluginsClient) ListAgentServicePlugins(ctx context.Context, in *empty.Empty, opts ...grpc.CallOption) (*AgentServicePluginList, error) {
	out := new(AgentServicePluginList)
	err := c.cc.Invoke(ctx, Plugins_ListAgentServicePlugins_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pluginsClient) InstallProvisionerPlugin(ctx context.Context, in *InstallPluginRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, Plugins_InstallProvisionerPlugin_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pluginsClient) InstallAgentServicePlugin(ctx context.Context, in *InstallPluginRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, Plugins_InstallAgentServicePlugin_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pluginsClient) UninstallProvisionerPlugin(ctx context.Context, in *UninstallPluginRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, Plugins_UninstallProvisionerPlugin_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pluginsClient) UninstallAgentServicePlugin(ctx context.Context, in *UninstallPluginRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	out := new(empty.Empty)
	err := c.cc.Invoke(ctx, Plugins_UninstallAgentServicePlugin_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PluginsServer is the server API for Plugins service.
// All implementations should embed UnimplementedPluginsServer
// for forward compatibility
type PluginsServer interface {
	ListProvisionerPlugins(context.Context, *empty.Empty) (*ProvisionerPluginList, error)
	ListAgentServicePlugins(context.Context, *empty.Empty) (*AgentServicePluginList, error)
	InstallProvisionerPlugin(context.Context, *InstallPluginRequest) (*empty.Empty, error)
	InstallAgentServicePlugin(context.Context, *InstallPluginRequest) (*empty.Empty, error)
	UninstallProvisionerPlugin(context.Context, *UninstallPluginRequest) (*empty.Empty, error)
	UninstallAgentServicePlugin(context.Context, *UninstallPluginRequest) (*empty.Empty, error)
}

// UnimplementedPluginsServer should be embedded to have forward compatible implementations.
type UnimplementedPluginsServer struct {
}

func (UnimplementedPluginsServer) ListProvisionerPlugins(context.Context, *empty.Empty) (*ProvisionerPluginList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListProvisionerPlugins not implemented")
}
func (UnimplementedPluginsServer) ListAgentServicePlugins(context.Context, *empty.Empty) (*AgentServicePluginList, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListAgentServicePlugins not implemented")
}
func (UnimplementedPluginsServer) InstallProvisionerPlugin(context.Context, *InstallPluginRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InstallProvisionerPlugin not implemented")
}
func (UnimplementedPluginsServer) InstallAgentServicePlugin(context.Context, *InstallPluginRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method InstallAgentServicePlugin not implemented")
}
func (UnimplementedPluginsServer) UninstallProvisionerPlugin(context.Context, *UninstallPluginRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UninstallProvisionerPlugin not implemented")
}
func (UnimplementedPluginsServer) UninstallAgentServicePlugin(context.Context, *UninstallPluginRequest) (*empty.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UninstallAgentServicePlugin not implemented")
}

// UnsafePluginsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PluginsServer will
// result in compilation errors.
type UnsafePluginsServer interface {
	mustEmbedUnimplementedPluginsServer()
}

func RegisterPluginsServer(s grpc.ServiceRegistrar, srv PluginsServer) {
	s.RegisterService(&Plugins_ServiceDesc, srv)
}

func _Plugins_ListProvisionerPlugins_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginsServer).ListProvisionerPlugins(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Plugins_ListProvisionerPlugins_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginsServer).ListProvisionerPlugins(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Plugins_ListAgentServicePlugins_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(empty.Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginsServer).ListAgentServicePlugins(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Plugins_ListAgentServicePlugins_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginsServer).ListAgentServicePlugins(ctx, req.(*empty.Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _Plugins_InstallProvisionerPlugin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InstallPluginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginsServer).InstallProvisionerPlugin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Plugins_InstallProvisionerPlugin_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginsServer).InstallProvisionerPlugin(ctx, req.(*InstallPluginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Plugins_InstallAgentServicePlugin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(InstallPluginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginsServer).InstallAgentServicePlugin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Plugins_InstallAgentServicePlugin_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginsServer).InstallAgentServicePlugin(ctx, req.(*InstallPluginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Plugins_UninstallProvisionerPlugin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UninstallPluginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginsServer).UninstallProvisionerPlugin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Plugins_UninstallProvisionerPlugin_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginsServer).UninstallProvisionerPlugin(ctx, req.(*UninstallPluginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Plugins_UninstallAgentServicePlugin_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UninstallPluginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginsServer).UninstallAgentServicePlugin(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: Plugins_UninstallAgentServicePlugin_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginsServer).UninstallAgentServicePlugin(ctx, req.(*UninstallPluginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Plugins_ServiceDesc is the grpc.ServiceDesc for Plugins service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Plugins_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Plugins",
	HandlerType: (*PluginsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListProvisionerPlugins",
			Handler:    _Plugins_ListProvisionerPlugins_Handler,
		},
		{
			MethodName: "ListAgentServicePlugins",
			Handler:    _Plugins_ListAgentServicePlugins_Handler,
		},
		{
			MethodName: "InstallProvisionerPlugin",
			Handler:    _Plugins_InstallProvisionerPlugin_Handler,
		},
		{
			MethodName: "InstallAgentServicePlugin",
			Handler:    _Plugins_InstallAgentServicePlugin_Handler,
		},
		{
			MethodName: "UninstallProvisionerPlugin",
			Handler:    _Plugins_UninstallProvisionerPlugin_Handler,
		},
		{
			MethodName: "UninstallAgentServicePlugin",
			Handler:    _Plugins_UninstallAgentServicePlugin_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "common/grpc/proto/plugins.proto",
}
