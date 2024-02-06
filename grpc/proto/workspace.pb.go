// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        v3.12.4
// source: grpc/proto/workspace.proto

package proto

import (
	types "github.com/daytonaio/daytona/grpc/proto/types"
	empty "github.com/golang/protobuf/ptypes/empty"
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

type CreateWorkspaceRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name         string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Repositories []string `protobuf:"bytes,2,rep,name=repositories,proto3" json:"repositories,omitempty"`
}

func (x *CreateWorkspaceRequest) Reset() {
	*x = CreateWorkspaceRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_proto_workspace_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateWorkspaceRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateWorkspaceRequest) ProtoMessage() {}

func (x *CreateWorkspaceRequest) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_proto_workspace_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateWorkspaceRequest.ProtoReflect.Descriptor instead.
func (*CreateWorkspaceRequest) Descriptor() ([]byte, []int) {
	return file_grpc_proto_workspace_proto_rawDescGZIP(), []int{0}
}

func (x *CreateWorkspaceRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *CreateWorkspaceRequest) GetRepositories() []string {
	if x != nil {
		return x.Repositories
	}
	return nil
}

type CreateWorkspaceResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Event   string `protobuf:"bytes,1,opt,name=event,proto3" json:"event,omitempty"`
	Payload string `protobuf:"bytes,2,opt,name=payload,proto3" json:"payload,omitempty"`
}

func (x *CreateWorkspaceResponse) Reset() {
	*x = CreateWorkspaceResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_proto_workspace_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateWorkspaceResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateWorkspaceResponse) ProtoMessage() {}

func (x *CreateWorkspaceResponse) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_proto_workspace_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateWorkspaceResponse.ProtoReflect.Descriptor instead.
func (*CreateWorkspaceResponse) Descriptor() ([]byte, []int) {
	return file_grpc_proto_workspace_proto_rawDescGZIP(), []int{1}
}

func (x *CreateWorkspaceResponse) GetEvent() string {
	if x != nil {
		return x.Event
	}
	return ""
}

func (x *CreateWorkspaceResponse) GetPayload() string {
	if x != nil {
		return x.Payload
	}
	return ""
}

type WorkspaceInfoRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *WorkspaceInfoRequest) Reset() {
	*x = WorkspaceInfoRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_proto_workspace_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WorkspaceInfoRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WorkspaceInfoRequest) ProtoMessage() {}

func (x *WorkspaceInfoRequest) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_proto_workspace_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WorkspaceInfoRequest.ProtoReflect.Descriptor instead.
func (*WorkspaceInfoRequest) Descriptor() ([]byte, []int) {
	return file_grpc_proto_workspace_proto_rawDescGZIP(), []int{2}
}

func (x *WorkspaceInfoRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type WorkspacePortForwardResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ContainerPort uint32 `protobuf:"varint,1,opt,name=containerPort,proto3" json:"containerPort,omitempty"`
	HostPort      uint32 `protobuf:"varint,2,opt,name=hostPort,proto3" json:"hostPort,omitempty"`
}

func (x *WorkspacePortForwardResponse) Reset() {
	*x = WorkspacePortForwardResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_proto_workspace_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WorkspacePortForwardResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WorkspacePortForwardResponse) ProtoMessage() {}

func (x *WorkspacePortForwardResponse) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_proto_workspace_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WorkspacePortForwardResponse.ProtoReflect.Descriptor instead.
func (*WorkspacePortForwardResponse) Descriptor() ([]byte, []int) {
	return file_grpc_proto_workspace_proto_rawDescGZIP(), []int{3}
}

func (x *WorkspacePortForwardResponse) GetContainerPort() uint32 {
	if x != nil {
		return x.ContainerPort
	}
	return 0
}

func (x *WorkspacePortForwardResponse) GetHostPort() uint32 {
	if x != nil {
		return x.HostPort
	}
	return 0
}

type WorkspaceStartRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id      string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Project string `protobuf:"bytes,2,opt,name=project,proto3" json:"project,omitempty"`
}

func (x *WorkspaceStartRequest) Reset() {
	*x = WorkspaceStartRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_proto_workspace_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WorkspaceStartRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WorkspaceStartRequest) ProtoMessage() {}

func (x *WorkspaceStartRequest) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_proto_workspace_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WorkspaceStartRequest.ProtoReflect.Descriptor instead.
func (*WorkspaceStartRequest) Descriptor() ([]byte, []int) {
	return file_grpc_proto_workspace_proto_rawDescGZIP(), []int{4}
}

func (x *WorkspaceStartRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *WorkspaceStartRequest) GetProject() string {
	if x != nil {
		return x.Project
	}
	return ""
}

type WorkspaceStopRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id      string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Project string `protobuf:"bytes,2,opt,name=project,proto3" json:"project,omitempty"`
}

func (x *WorkspaceStopRequest) Reset() {
	*x = WorkspaceStopRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_proto_workspace_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WorkspaceStopRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WorkspaceStopRequest) ProtoMessage() {}

func (x *WorkspaceStopRequest) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_proto_workspace_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WorkspaceStopRequest.ProtoReflect.Descriptor instead.
func (*WorkspaceStopRequest) Descriptor() ([]byte, []int) {
	return file_grpc_proto_workspace_proto_rawDescGZIP(), []int{5}
}

func (x *WorkspaceStopRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *WorkspaceStopRequest) GetProject() string {
	if x != nil {
		return x.Project
	}
	return ""
}

type WorkspaceRemoveRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *WorkspaceRemoveRequest) Reset() {
	*x = WorkspaceRemoveRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_proto_workspace_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WorkspaceRemoveRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WorkspaceRemoveRequest) ProtoMessage() {}

func (x *WorkspaceRemoveRequest) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_proto_workspace_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WorkspaceRemoveRequest.ProtoReflect.Descriptor instead.
func (*WorkspaceRemoveRequest) Descriptor() ([]byte, []int) {
	return file_grpc_proto_workspace_proto_rawDescGZIP(), []int{6}
}

func (x *WorkspaceRemoveRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type WorkspaceListResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Workspaces []*types.WorkspaceInfo `protobuf:"bytes,1,rep,name=workspaces,proto3" json:"workspaces,omitempty"`
}

func (x *WorkspaceListResponse) Reset() {
	*x = WorkspaceListResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_grpc_proto_workspace_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WorkspaceListResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WorkspaceListResponse) ProtoMessage() {}

func (x *WorkspaceListResponse) ProtoReflect() protoreflect.Message {
	mi := &file_grpc_proto_workspace_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WorkspaceListResponse.ProtoReflect.Descriptor instead.
func (*WorkspaceListResponse) Descriptor() ([]byte, []int) {
	return file_grpc_proto_workspace_proto_rawDescGZIP(), []int{7}
}

func (x *WorkspaceListResponse) GetWorkspaces() []*types.WorkspaceInfo {
	if x != nil {
		return x.Workspaces
	}
	return nil
}

var File_grpc_proto_workspace_proto protoreflect.FileDescriptor

var file_grpc_proto_workspace_proto_rawDesc = []byte{
	0x0a, 0x1a, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x77, 0x6f, 0x72,
	0x6b, 0x73, 0x70, 0x61, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1b, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d,
	0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x72, 0x70, 0x63, 0x2f,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x74, 0x79, 0x70, 0x65,
	0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x50, 0x0a, 0x16, 0x43, 0x72, 0x65, 0x61, 0x74,
	0x65, 0x57, 0x6f, 0x72, 0x6b, 0x73, 0x70, 0x61, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x22, 0x0a, 0x0c, 0x72, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74,
	0x6f, 0x72, 0x69, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0c, 0x72, 0x65, 0x70,
	0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x69, 0x65, 0x73, 0x22, 0x49, 0x0a, 0x17, 0x43, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x57, 0x6f, 0x72, 0x6b, 0x73, 0x70, 0x61, 0x63, 0x65, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x61,
	0x79, 0x6c, 0x6f, 0x61, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x70, 0x61, 0x79,
	0x6c, 0x6f, 0x61, 0x64, 0x22, 0x26, 0x0a, 0x14, 0x57, 0x6f, 0x72, 0x6b, 0x73, 0x70, 0x61, 0x63,
	0x65, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0x60, 0x0a, 0x1c,
	0x57, 0x6f, 0x72, 0x6b, 0x73, 0x70, 0x61, 0x63, 0x65, 0x50, 0x6f, 0x72, 0x74, 0x46, 0x6f, 0x72,
	0x77, 0x61, 0x72, 0x64, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x24, 0x0a, 0x0d,
	0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x50, 0x6f, 0x72, 0x74, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0d, 0x52, 0x0d, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x50, 0x6f,
	0x72, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x68, 0x6f, 0x73, 0x74, 0x50, 0x6f, 0x72, 0x74, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x08, 0x68, 0x6f, 0x73, 0x74, 0x50, 0x6f, 0x72, 0x74, 0x22, 0x41,
	0x0a, 0x15, 0x57, 0x6f, 0x72, 0x6b, 0x73, 0x70, 0x61, 0x63, 0x65, 0x53, 0x74, 0x61, 0x72, 0x74,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x72, 0x6f, 0x6a, 0x65,
	0x63, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x70, 0x72, 0x6f, 0x6a, 0x65, 0x63,
	0x74, 0x22, 0x40, 0x0a, 0x14, 0x57, 0x6f, 0x72, 0x6b, 0x73, 0x70, 0x61, 0x63, 0x65, 0x53, 0x74,
	0x6f, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x72, 0x6f,
	0x6a, 0x65, 0x63, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x70, 0x72, 0x6f, 0x6a,
	0x65, 0x63, 0x74, 0x22, 0x28, 0x0a, 0x16, 0x57, 0x6f, 0x72, 0x6b, 0x73, 0x70, 0x61, 0x63, 0x65,
	0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a,
	0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0x47, 0x0a,
	0x15, 0x57, 0x6f, 0x72, 0x6b, 0x73, 0x70, 0x61, 0x63, 0x65, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2e, 0x0a, 0x0a, 0x77, 0x6f, 0x72, 0x6b, 0x73, 0x70,
	0x61, 0x63, 0x65, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x57, 0x6f, 0x72,
	0x6b, 0x73, 0x70, 0x61, 0x63, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x0a, 0x77, 0x6f, 0x72, 0x6b,
	0x73, 0x70, 0x61, 0x63, 0x65, 0x73, 0x32, 0xef, 0x02, 0x0a, 0x10, 0x57, 0x6f, 0x72, 0x6b, 0x73,
	0x70, 0x61, 0x63, 0x65, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x3f, 0x0a, 0x06, 0x43,
	0x72, 0x65, 0x61, 0x74, 0x65, 0x12, 0x17, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x57, 0x6f,
	0x72, 0x6b, 0x73, 0x70, 0x61, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x18,
	0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x57, 0x6f, 0x72, 0x6b, 0x73, 0x70, 0x61, 0x63, 0x65,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x30, 0x01, 0x12, 0x2f, 0x0a, 0x04,
	0x49, 0x6e, 0x66, 0x6f, 0x12, 0x15, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x73, 0x70, 0x61, 0x63, 0x65,
	0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0e, 0x2e, 0x57, 0x6f,
	0x72, 0x6b, 0x73, 0x70, 0x61, 0x63, 0x65, 0x49, 0x6e, 0x66, 0x6f, 0x22, 0x00, 0x12, 0x38, 0x0a,
	0x04, 0x4c, 0x69, 0x73, 0x74, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x16, 0x2e,
	0x57, 0x6f, 0x72, 0x6b, 0x73, 0x70, 0x61, 0x63, 0x65, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x39, 0x0a, 0x05, 0x53, 0x74, 0x61, 0x72, 0x74,
	0x12, 0x16, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x73, 0x70, 0x61, 0x63, 0x65, 0x53, 0x74, 0x61, 0x72,
	0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79,
	0x22, 0x00, 0x12, 0x37, 0x0a, 0x04, 0x53, 0x74, 0x6f, 0x70, 0x12, 0x15, 0x2e, 0x57, 0x6f, 0x72,
	0x6b, 0x73, 0x70, 0x61, 0x63, 0x65, 0x53, 0x74, 0x6f, 0x70, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73,
	0x74, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x12, 0x3b, 0x0a, 0x06, 0x52,
	0x65, 0x6d, 0x6f, 0x76, 0x65, 0x12, 0x17, 0x2e, 0x57, 0x6f, 0x72, 0x6b, 0x73, 0x70, 0x61, 0x63,
	0x65, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x42, 0x09, 0x5a, 0x07, 0x2e, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_grpc_proto_workspace_proto_rawDescOnce sync.Once
	file_grpc_proto_workspace_proto_rawDescData = file_grpc_proto_workspace_proto_rawDesc
)

func file_grpc_proto_workspace_proto_rawDescGZIP() []byte {
	file_grpc_proto_workspace_proto_rawDescOnce.Do(func() {
		file_grpc_proto_workspace_proto_rawDescData = protoimpl.X.CompressGZIP(file_grpc_proto_workspace_proto_rawDescData)
	})
	return file_grpc_proto_workspace_proto_rawDescData
}

var file_grpc_proto_workspace_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_grpc_proto_workspace_proto_goTypes = []interface{}{
	(*CreateWorkspaceRequest)(nil),       // 0: CreateWorkspaceRequest
	(*CreateWorkspaceResponse)(nil),      // 1: CreateWorkspaceResponse
	(*WorkspaceInfoRequest)(nil),         // 2: WorkspaceInfoRequest
	(*WorkspacePortForwardResponse)(nil), // 3: WorkspacePortForwardResponse
	(*WorkspaceStartRequest)(nil),        // 4: WorkspaceStartRequest
	(*WorkspaceStopRequest)(nil),         // 5: WorkspaceStopRequest
	(*WorkspaceRemoveRequest)(nil),       // 6: WorkspaceRemoveRequest
	(*WorkspaceListResponse)(nil),        // 7: WorkspaceListResponse
	(*types.WorkspaceInfo)(nil),          // 8: WorkspaceInfo
	(*empty.Empty)(nil),                  // 9: google.protobuf.Empty
}
var file_grpc_proto_workspace_proto_depIdxs = []int32{
	8, // 0: WorkspaceListResponse.workspaces:type_name -> WorkspaceInfo
	0, // 1: WorkspaceService.Create:input_type -> CreateWorkspaceRequest
	2, // 2: WorkspaceService.Info:input_type -> WorkspaceInfoRequest
	9, // 3: WorkspaceService.List:input_type -> google.protobuf.Empty
	4, // 4: WorkspaceService.Start:input_type -> WorkspaceStartRequest
	5, // 5: WorkspaceService.Stop:input_type -> WorkspaceStopRequest
	6, // 6: WorkspaceService.Remove:input_type -> WorkspaceRemoveRequest
	1, // 7: WorkspaceService.Create:output_type -> CreateWorkspaceResponse
	8, // 8: WorkspaceService.Info:output_type -> WorkspaceInfo
	7, // 9: WorkspaceService.List:output_type -> WorkspaceListResponse
	9, // 10: WorkspaceService.Start:output_type -> google.protobuf.Empty
	9, // 11: WorkspaceService.Stop:output_type -> google.protobuf.Empty
	9, // 12: WorkspaceService.Remove:output_type -> google.protobuf.Empty
	7, // [7:13] is the sub-list for method output_type
	1, // [1:7] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_grpc_proto_workspace_proto_init() }
func file_grpc_proto_workspace_proto_init() {
	if File_grpc_proto_workspace_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_grpc_proto_workspace_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateWorkspaceRequest); i {
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
		file_grpc_proto_workspace_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateWorkspaceResponse); i {
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
		file_grpc_proto_workspace_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WorkspaceInfoRequest); i {
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
		file_grpc_proto_workspace_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WorkspacePortForwardResponse); i {
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
		file_grpc_proto_workspace_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WorkspaceStartRequest); i {
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
		file_grpc_proto_workspace_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WorkspaceStopRequest); i {
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
		file_grpc_proto_workspace_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WorkspaceRemoveRequest); i {
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
		file_grpc_proto_workspace_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WorkspaceListResponse); i {
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
			RawDescriptor: file_grpc_proto_workspace_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_grpc_proto_workspace_proto_goTypes,
		DependencyIndexes: file_grpc_proto_workspace_proto_depIdxs,
		MessageInfos:      file_grpc_proto_workspace_proto_msgTypes,
	}.Build()
	File_grpc_proto_workspace_proto = out.File
	file_grpc_proto_workspace_proto_rawDesc = nil
	file_grpc_proto_workspace_proto_goTypes = nil
	file_grpc_proto_workspace_proto_depIdxs = nil
}
