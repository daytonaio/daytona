// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        v3.12.4
// source: common/grpc/proto/server.proto

package proto

import (
	types "github.com/daytonaio/daytona/common/grpc/proto/types"
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

type GetPublicKeyResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PublicKey string `protobuf:"bytes,1,opt,name=publicKey,proto3" json:"publicKey,omitempty"`
}

func (x *GetPublicKeyResponse) Reset() {
	*x = GetPublicKeyResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_grpc_proto_server_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *GetPublicKeyResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetPublicKeyResponse) ProtoMessage() {}

func (x *GetPublicKeyResponse) ProtoReflect() protoreflect.Message {
	mi := &file_common_grpc_proto_server_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetPublicKeyResponse.ProtoReflect.Descriptor instead.
func (*GetPublicKeyResponse) Descriptor() ([]byte, []int) {
	return file_common_grpc_proto_server_proto_rawDescGZIP(), []int{0}
}

func (x *GetPublicKeyResponse) GetPublicKey() string {
	if x != nil {
		return x.PublicKey
	}
	return ""
}

type SetKeyRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	PrivateKey string `protobuf:"bytes,1,opt,name=privateKey,proto3" json:"privateKey,omitempty"`
}

func (x *SetKeyRequest) Reset() {
	*x = SetKeyRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_grpc_proto_server_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SetKeyRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SetKeyRequest) ProtoMessage() {}

func (x *SetKeyRequest) ProtoReflect() protoreflect.Message {
	mi := &file_common_grpc_proto_server_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SetKeyRequest.ProtoReflect.Descriptor instead.
func (*SetKeyRequest) Descriptor() ([]byte, []int) {
	return file_common_grpc_proto_server_proto_rawDescGZIP(), []int{1}
}

func (x *SetKeyRequest) GetPrivateKey() string {
	if x != nil {
		return x.PrivateKey
	}
	return ""
}

type AuthKey struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
}

func (x *AuthKey) Reset() {
	*x = AuthKey{}
	if protoimpl.UnsafeEnabled {
		mi := &file_common_grpc_proto_server_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AuthKey) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AuthKey) ProtoMessage() {}

func (x *AuthKey) ProtoReflect() protoreflect.Message {
	mi := &file_common_grpc_proto_server_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AuthKey.ProtoReflect.Descriptor instead.
func (*AuthKey) Descriptor() ([]byte, []int) {
	return file_common_grpc_proto_server_proto_rawDescGZIP(), []int{2}
}

func (x *AuthKey) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

var File_common_grpc_proto_server_proto protoreflect.FileDescriptor

var file_common_grpc_proto_server_proto_rawDesc = []byte{
	0x0a, 0x1e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x65, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x23, 0x63,
	0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2f, 0x67, 0x72, 0x70, 0x63, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x22, 0x34, 0x0a, 0x14, 0x47, 0x65, 0x74, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b,
	0x65, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x70, 0x75,
	0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x70,
	0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x22, 0x2f, 0x0a, 0x0d, 0x53, 0x65, 0x74, 0x4b,
	0x65, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x70, 0x72, 0x69,
	0x76, 0x61, 0x74, 0x65, 0x4b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x70,
	0x72, 0x69, 0x76, 0x61, 0x74, 0x65, 0x4b, 0x65, 0x79, 0x22, 0x1b, 0x0a, 0x07, 0x41, 0x75, 0x74,
	0x68, 0x4b, 0x65, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x32, 0x9e, 0x03, 0x0a, 0x06, 0x53, 0x65, 0x72, 0x76, 0x65,
	0x72, 0x12, 0x3f, 0x0a, 0x0c, 0x47, 0x65, 0x74, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65,
	0x79, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x15, 0x2e, 0x47, 0x65, 0x74, 0x50,
	0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x00, 0x12, 0x3e, 0x0a, 0x0b, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x4b, 0x65,
	0x79, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x15, 0x2e, 0x47, 0x65, 0x74, 0x50,
	0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65,
	0x22, 0x00, 0x12, 0x31, 0x0a, 0x06, 0x53, 0x65, 0x74, 0x4b, 0x65, 0x79, 0x12, 0x0e, 0x2e, 0x53,
	0x65, 0x74, 0x4b, 0x65, 0x79, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x15, 0x2e, 0x47,
	0x65, 0x74, 0x50, 0x75, 0x62, 0x6c, 0x69, 0x63, 0x4b, 0x65, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x3d, 0x0a, 0x09, 0x44, 0x65, 0x6c, 0x65, 0x74, 0x65, 0x4b,
	0x65, 0x79, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f,
	0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70,
	0x74, 0x79, 0x22, 0x00, 0x12, 0x34, 0x0a, 0x09, 0x47, 0x65, 0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69,
	0x67, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x0d, 0x2e, 0x53, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x22, 0x00, 0x12, 0x34, 0x0a, 0x09, 0x53, 0x65,
	0x74, 0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x12, 0x0d, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72,
	0x43, 0x6f, 0x6e, 0x66, 0x69, 0x67, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00,
	0x12, 0x35, 0x0a, 0x0f, 0x47, 0x65, 0x6e, 0x65, 0x72, 0x61, 0x74, 0x65, 0x41, 0x75, 0x74, 0x68,
	0x4b, 0x65, 0x79, 0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x08, 0x2e, 0x41, 0x75,
	0x74, 0x68, 0x4b, 0x65, 0x79, 0x22, 0x00, 0x42, 0x09, 0x5a, 0x07, 0x2e, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_common_grpc_proto_server_proto_rawDescOnce sync.Once
	file_common_grpc_proto_server_proto_rawDescData = file_common_grpc_proto_server_proto_rawDesc
)

func file_common_grpc_proto_server_proto_rawDescGZIP() []byte {
	file_common_grpc_proto_server_proto_rawDescOnce.Do(func() {
		file_common_grpc_proto_server_proto_rawDescData = protoimpl.X.CompressGZIP(file_common_grpc_proto_server_proto_rawDescData)
	})
	return file_common_grpc_proto_server_proto_rawDescData
}

var file_common_grpc_proto_server_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_common_grpc_proto_server_proto_goTypes = []interface{}{
	(*GetPublicKeyResponse)(nil), // 0: GetPublicKeyResponse
	(*SetKeyRequest)(nil),        // 1: SetKeyRequest
	(*AuthKey)(nil),              // 2: AuthKey
	(*empty.Empty)(nil),          // 3: google.protobuf.Empty
	(*types.ServerConfig)(nil),   // 4: ServerConfig
}
var file_common_grpc_proto_server_proto_depIdxs = []int32{
	3, // 0: Server.GetPublicKey:input_type -> google.protobuf.Empty
	3, // 1: Server.GenerateKey:input_type -> google.protobuf.Empty
	1, // 2: Server.SetKey:input_type -> SetKeyRequest
	3, // 3: Server.DeleteKey:input_type -> google.protobuf.Empty
	3, // 4: Server.GetConfig:input_type -> google.protobuf.Empty
	4, // 5: Server.SetConfig:input_type -> ServerConfig
	3, // 6: Server.GenerateAuthKey:input_type -> google.protobuf.Empty
	0, // 7: Server.GetPublicKey:output_type -> GetPublicKeyResponse
	0, // 8: Server.GenerateKey:output_type -> GetPublicKeyResponse
	0, // 9: Server.SetKey:output_type -> GetPublicKeyResponse
	3, // 10: Server.DeleteKey:output_type -> google.protobuf.Empty
	4, // 11: Server.GetConfig:output_type -> ServerConfig
	3, // 12: Server.SetConfig:output_type -> google.protobuf.Empty
	2, // 13: Server.GenerateAuthKey:output_type -> AuthKey
	7, // [7:14] is the sub-list for method output_type
	0, // [0:7] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_common_grpc_proto_server_proto_init() }
func file_common_grpc_proto_server_proto_init() {
	if File_common_grpc_proto_server_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_common_grpc_proto_server_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*GetPublicKeyResponse); i {
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
		file_common_grpc_proto_server_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SetKeyRequest); i {
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
		file_common_grpc_proto_server_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AuthKey); i {
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
			RawDescriptor: file_common_grpc_proto_server_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_common_grpc_proto_server_proto_goTypes,
		DependencyIndexes: file_common_grpc_proto_server_proto_depIdxs,
		MessageInfos:      file_common_grpc_proto_server_proto_msgTypes,
	}.Build()
	File_common_grpc_proto_server_proto = out.File
	file_common_grpc_proto_server_proto_rawDesc = nil
	file_common_grpc_proto_server_proto_goTypes = nil
	file_common_grpc_proto_server_proto_depIdxs = nil
}
