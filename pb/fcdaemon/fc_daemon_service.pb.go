// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.15.2
// source: fc_daemon_service.proto

package fcdaemon

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type CreateRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	KernelImagePath    string `protobuf:"bytes,4,opt,name=kernelImagePath,proto3" json:"kernelImagePath,omitempty"`
	KernelArgs         string `protobuf:"bytes,5,opt,name=kernelArgs,proto3" json:"kernelArgs,omitempty"`
	InitrdPath         string `protobuf:"bytes,17,opt,name=initrdPath,proto3" json:"initrdPath,omitempty"`
	ImageName          string `protobuf:"bytes,6,opt,name=imageName,proto3" json:"imageName,omitempty"`
	ImageStoragePrefix string `protobuf:"bytes,7,opt,name=imageStoragePrefix,proto3" json:"imageStoragePrefix,omitempty"`
	InetDev            string `protobuf:"bytes,8,opt,name=inetDev,proto3" json:"inetDev,omitempty"`
	NameServer         string `protobuf:"bytes,9,opt,name=nameServer,proto3" json:"nameServer,omitempty"`
	AwsAccessKeyID     string `protobuf:"bytes,10,opt,name=awsAccessKeyID,proto3" json:"awsAccessKeyID,omitempty"`
	AwsSecretAccessKey string `protobuf:"bytes,11,opt,name=awsSecretAccessKey,proto3" json:"awsSecretAccessKey,omitempty"`
	Cpu                int64  `protobuf:"varint,13,opt,name=cpu,proto3" json:"cpu,omitempty"`
	Mem                int64  `protobuf:"varint,14,opt,name=mem,proto3" json:"mem,omitempty"`
	NumReplica         uint32 `protobuf:"varint,15,opt,name=numReplica,proto3" json:"numReplica,omitempty"`
	VmType             string `protobuf:"bytes,16,opt,name=vmType,proto3" json:"vmType,omitempty"`
}

func (x *CreateRequest) Reset() {
	*x = CreateRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_fc_daemon_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateRequest) ProtoMessage() {}

func (x *CreateRequest) ProtoReflect() protoreflect.Message {
	mi := &file_fc_daemon_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateRequest.ProtoReflect.Descriptor instead.
func (*CreateRequest) Descriptor() ([]byte, []int) {
	return file_fc_daemon_service_proto_rawDescGZIP(), []int{0}
}

func (x *CreateRequest) GetKernelImagePath() string {
	if x != nil {
		return x.KernelImagePath
	}
	return ""
}

func (x *CreateRequest) GetKernelArgs() string {
	if x != nil {
		return x.KernelArgs
	}
	return ""
}

func (x *CreateRequest) GetInitrdPath() string {
	if x != nil {
		return x.InitrdPath
	}
	return ""
}

func (x *CreateRequest) GetImageName() string {
	if x != nil {
		return x.ImageName
	}
	return ""
}

func (x *CreateRequest) GetImageStoragePrefix() string {
	if x != nil {
		return x.ImageStoragePrefix
	}
	return ""
}

func (x *CreateRequest) GetInetDev() string {
	if x != nil {
		return x.InetDev
	}
	return ""
}

func (x *CreateRequest) GetNameServer() string {
	if x != nil {
		return x.NameServer
	}
	return ""
}

func (x *CreateRequest) GetAwsAccessKeyID() string {
	if x != nil {
		return x.AwsAccessKeyID
	}
	return ""
}

func (x *CreateRequest) GetAwsSecretAccessKey() string {
	if x != nil {
		return x.AwsSecretAccessKey
	}
	return ""
}

func (x *CreateRequest) GetCpu() int64 {
	if x != nil {
		return x.Cpu
	}
	return 0
}

func (x *CreateRequest) GetMem() int64 {
	if x != nil {
		return x.Mem
	}
	return 0
}

func (x *CreateRequest) GetNumReplica() uint32 {
	if x != nil {
		return x.NumReplica
	}
	return 0
}

func (x *CreateRequest) GetVmType() string {
	if x != nil {
		return x.VmType
	}
	return ""
}

type CreateReply struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ip   string `protobuf:"bytes,1,opt,name=ip,proto3" json:"ip,omitempty"`
	Vmid string `protobuf:"bytes,2,opt,name=vmid,proto3" json:"vmid,omitempty"`
}

func (x *CreateReply) Reset() {
	*x = CreateReply{}
	if protoimpl.UnsafeEnabled {
		mi := &file_fc_daemon_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CreateReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateReply) ProtoMessage() {}

func (x *CreateReply) ProtoReflect() protoreflect.Message {
	mi := &file_fc_daemon_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateReply.ProtoReflect.Descriptor instead.
func (*CreateReply) Descriptor() ([]byte, []int) {
	return file_fc_daemon_service_proto_rawDescGZIP(), []int{1}
}

func (x *CreateReply) GetIp() string {
	if x != nil {
		return x.Ip
	}
	return ""
}

func (x *CreateReply) GetVmid() string {
	if x != nil {
		return x.Vmid
	}
	return ""
}

type ReleaseRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Vmid string `protobuf:"bytes,1,opt,name=vmid,proto3" json:"vmid,omitempty"`
}

func (x *ReleaseRequest) Reset() {
	*x = ReleaseRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_fc_daemon_service_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ReleaseRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ReleaseRequest) ProtoMessage() {}

func (x *ReleaseRequest) ProtoReflect() protoreflect.Message {
	mi := &file_fc_daemon_service_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ReleaseRequest.ProtoReflect.Descriptor instead.
func (*ReleaseRequest) Descriptor() ([]byte, []int) {
	return file_fc_daemon_service_proto_rawDescGZIP(), []int{2}
}

func (x *ReleaseRequest) GetVmid() string {
	if x != nil {
		return x.Vmid
	}
	return ""
}

type CpuHogs struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	N int32 `protobuf:"varint,1,opt,name=n,proto3" json:"n,omitempty"`
}

func (x *CpuHogs) Reset() {
	*x = CpuHogs{}
	if protoimpl.UnsafeEnabled {
		mi := &file_fc_daemon_service_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CpuHogs) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CpuHogs) ProtoMessage() {}

func (x *CpuHogs) ProtoReflect() protoreflect.Message {
	mi := &file_fc_daemon_service_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CpuHogs.ProtoReflect.Descriptor instead.
func (*CpuHogs) Descriptor() ([]byte, []int) {
	return file_fc_daemon_service_proto_rawDescGZIP(), []int{3}
}

func (x *CpuHogs) GetN() int32 {
	if x != nil {
		return x.N
	}
	return 0
}

var File_fc_daemon_service_proto protoreflect.FileDescriptor

var file_fc_daemon_service_proto_rawDesc = []byte{
	0x0a, 0x17, 0x66, 0x63, 0x5f, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x5f, 0x73, 0x65, 0x72, 0x76,
	0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x66, 0x63, 0x64, 0x61, 0x65,
	0x6d, 0x6f, 0x6e, 0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0xb5, 0x03, 0x0a, 0x0d, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x12, 0x28, 0x0a, 0x0f, 0x6b, 0x65, 0x72, 0x6e, 0x65, 0x6c, 0x49, 0x6d, 0x61, 0x67,
	0x65, 0x50, 0x61, 0x74, 0x68, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0f, 0x6b, 0x65, 0x72,
	0x6e, 0x65, 0x6c, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x50, 0x61, 0x74, 0x68, 0x12, 0x1e, 0x0a, 0x0a,
	0x6b, 0x65, 0x72, 0x6e, 0x65, 0x6c, 0x41, 0x72, 0x67, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0a, 0x6b, 0x65, 0x72, 0x6e, 0x65, 0x6c, 0x41, 0x72, 0x67, 0x73, 0x12, 0x1e, 0x0a, 0x0a,
	0x69, 0x6e, 0x69, 0x74, 0x72, 0x64, 0x50, 0x61, 0x74, 0x68, 0x18, 0x11, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0a, 0x69, 0x6e, 0x69, 0x74, 0x72, 0x64, 0x50, 0x61, 0x74, 0x68, 0x12, 0x1c, 0x0a, 0x09,
	0x69, 0x6d, 0x61, 0x67, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x09, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x2e, 0x0a, 0x12, 0x69, 0x6d,
	0x61, 0x67, 0x65, 0x53, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x50, 0x72, 0x65, 0x66, 0x69, 0x78,
	0x18, 0x07, 0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x69, 0x6d, 0x61, 0x67, 0x65, 0x53, 0x74, 0x6f,
	0x72, 0x61, 0x67, 0x65, 0x50, 0x72, 0x65, 0x66, 0x69, 0x78, 0x12, 0x18, 0x0a, 0x07, 0x69, 0x6e,
	0x65, 0x74, 0x44, 0x65, 0x76, 0x18, 0x08, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x69, 0x6e, 0x65,
	0x74, 0x44, 0x65, 0x76, 0x12, 0x1e, 0x0a, 0x0a, 0x6e, 0x61, 0x6d, 0x65, 0x53, 0x65, 0x72, 0x76,
	0x65, 0x72, 0x18, 0x09, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x6e, 0x61, 0x6d, 0x65, 0x53, 0x65,
	0x72, 0x76, 0x65, 0x72, 0x12, 0x26, 0x0a, 0x0e, 0x61, 0x77, 0x73, 0x41, 0x63, 0x63, 0x65, 0x73,
	0x73, 0x4b, 0x65, 0x79, 0x49, 0x44, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x61, 0x77,
	0x73, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73, 0x4b, 0x65, 0x79, 0x49, 0x44, 0x12, 0x2e, 0x0a, 0x12,
	0x61, 0x77, 0x73, 0x53, 0x65, 0x63, 0x72, 0x65, 0x74, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73, 0x4b,
	0x65, 0x79, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x09, 0x52, 0x12, 0x61, 0x77, 0x73, 0x53, 0x65, 0x63,
	0x72, 0x65, 0x74, 0x41, 0x63, 0x63, 0x65, 0x73, 0x73, 0x4b, 0x65, 0x79, 0x12, 0x10, 0x0a, 0x03,
	0x63, 0x70, 0x75, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x63, 0x70, 0x75, 0x12, 0x10,
	0x0a, 0x03, 0x6d, 0x65, 0x6d, 0x18, 0x0e, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x6d, 0x65, 0x6d,
	0x12, 0x1e, 0x0a, 0x0a, 0x6e, 0x75, 0x6d, 0x52, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x18, 0x0f,
	0x20, 0x01, 0x28, 0x0d, 0x52, 0x0a, 0x6e, 0x75, 0x6d, 0x52, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61,
	0x12, 0x16, 0x0a, 0x06, 0x76, 0x6d, 0x54, 0x79, 0x70, 0x65, 0x18, 0x10, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x06, 0x76, 0x6d, 0x54, 0x79, 0x70, 0x65, 0x22, 0x31, 0x0a, 0x0b, 0x43, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x70, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x70, 0x12, 0x12, 0x0a, 0x04, 0x76, 0x6d, 0x69, 0x64, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x76, 0x6d, 0x69, 0x64, 0x22, 0x24, 0x0a, 0x0e, 0x52,
	0x65, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a,
	0x04, 0x76, 0x6d, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x76, 0x6d, 0x69,
	0x64, 0x22, 0x17, 0x0a, 0x07, 0x43, 0x70, 0x75, 0x48, 0x6f, 0x67, 0x73, 0x12, 0x0c, 0x0a, 0x01,
	0x6e, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x01, 0x6e, 0x32, 0x8e, 0x02, 0x0a, 0x09, 0x46,
	0x63, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x41, 0x0a, 0x0d, 0x43, 0x72, 0x65, 0x61,
	0x74, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x12, 0x17, 0x2e, 0x66, 0x63, 0x64, 0x61,
	0x65, 0x6d, 0x6f, 0x6e, 0x2e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x15, 0x2e, 0x66, 0x63, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x2e, 0x43, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x22, 0x00, 0x12, 0x44, 0x0a, 0x0e, 0x52,
	0x65, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x56, 0x4d, 0x53, 0x65, 0x74, 0x75, 0x70, 0x12, 0x18, 0x2e,
	0x66, 0x63, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x2e, 0x52, 0x65, 0x6c, 0x65, 0x61, 0x73, 0x65,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22,
	0x00, 0x12, 0x3c, 0x0a, 0x0d, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x43, 0x50, 0x55, 0x48, 0x6f,
	0x67, 0x73, 0x12, 0x11, 0x2e, 0x66, 0x63, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x2e, 0x43, 0x70,
	0x75, 0x48, 0x6f, 0x67, 0x73, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x12,
	0x3a, 0x0a, 0x0b, 0x53, 0x74, 0x6f, 0x70, 0x43, 0x50, 0x55, 0x48, 0x6f, 0x67, 0x73, 0x12, 0x11,
	0x2e, 0x66, 0x63, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x2e, 0x43, 0x70, 0x75, 0x48, 0x6f, 0x67,
	0x73, 0x1a, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x22, 0x00, 0x42, 0x39, 0x0a, 0x19, 0x69,
	0x6f, 0x2e, 0x67, 0x72, 0x70, 0x63, 0x2e, 0x65, 0x78, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x73, 0x2e,
	0x66, 0x63, 0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x42, 0x0d, 0x46, 0x63, 0x44, 0x61, 0x65, 0x6d,
	0x6f, 0x6e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x0b, 0x2e, 0x2f, 0x3b, 0x66, 0x63,
	0x64, 0x61, 0x65, 0x6d, 0x6f, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_fc_daemon_service_proto_rawDescOnce sync.Once
	file_fc_daemon_service_proto_rawDescData = file_fc_daemon_service_proto_rawDesc
)

func file_fc_daemon_service_proto_rawDescGZIP() []byte {
	file_fc_daemon_service_proto_rawDescOnce.Do(func() {
		file_fc_daemon_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_fc_daemon_service_proto_rawDescData)
	})
	return file_fc_daemon_service_proto_rawDescData
}

var file_fc_daemon_service_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_fc_daemon_service_proto_goTypes = []interface{}{
	(*CreateRequest)(nil),  // 0: fcdaemon.CreateRequest
	(*CreateReply)(nil),    // 1: fcdaemon.CreateReply
	(*ReleaseRequest)(nil), // 2: fcdaemon.ReleaseRequest
	(*CpuHogs)(nil),        // 3: fcdaemon.CpuHogs
	(*emptypb.Empty)(nil),  // 4: google.protobuf.Empty
}
var file_fc_daemon_service_proto_depIdxs = []int32{
	0, // 0: fcdaemon.FcService.CreateReplica:input_type -> fcdaemon.CreateRequest
	2, // 1: fcdaemon.FcService.ReleaseVMSetup:input_type -> fcdaemon.ReleaseRequest
	3, // 2: fcdaemon.FcService.CreateCPUHogs:input_type -> fcdaemon.CpuHogs
	3, // 3: fcdaemon.FcService.StopCPUHogs:input_type -> fcdaemon.CpuHogs
	1, // 4: fcdaemon.FcService.CreateReplica:output_type -> fcdaemon.CreateReply
	4, // 5: fcdaemon.FcService.ReleaseVMSetup:output_type -> google.protobuf.Empty
	4, // 6: fcdaemon.FcService.CreateCPUHogs:output_type -> google.protobuf.Empty
	4, // 7: fcdaemon.FcService.StopCPUHogs:output_type -> google.protobuf.Empty
	4, // [4:8] is the sub-list for method output_type
	0, // [0:4] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_fc_daemon_service_proto_init() }
func file_fc_daemon_service_proto_init() {
	if File_fc_daemon_service_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_fc_daemon_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateRequest); i {
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
		file_fc_daemon_service_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CreateReply); i {
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
		file_fc_daemon_service_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ReleaseRequest); i {
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
		file_fc_daemon_service_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CpuHogs); i {
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
			RawDescriptor: file_fc_daemon_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_fc_daemon_service_proto_goTypes,
		DependencyIndexes: file_fc_daemon_service_proto_depIdxs,
		MessageInfos:      file_fc_daemon_service_proto_msgTypes,
	}.Build()
	File_fc_daemon_service_proto = out.File
	file_fc_daemon_service_proto_rawDesc = nil
	file_fc_daemon_service_proto_goTypes = nil
	file_fc_daemon_service_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// FcServiceClient is the client API for FcService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type FcServiceClient interface {
	// Asks the fcdaemon to create ONE replica
	CreateReplica(ctx context.Context, in *CreateRequest, opts ...grpc.CallOption) (*CreateReply, error)
	ReleaseVMSetup(ctx context.Context, in *ReleaseRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
	CreateCPUHogs(ctx context.Context, in *CpuHogs, opts ...grpc.CallOption) (*emptypb.Empty, error)
	StopCPUHogs(ctx context.Context, in *CpuHogs, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type fcServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewFcServiceClient(cc grpc.ClientConnInterface) FcServiceClient {
	return &fcServiceClient{cc}
}

func (c *fcServiceClient) CreateReplica(ctx context.Context, in *CreateRequest, opts ...grpc.CallOption) (*CreateReply, error) {
	out := new(CreateReply)
	err := c.cc.Invoke(ctx, "/fcdaemon.FcService/CreateReplica", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fcServiceClient) ReleaseVMSetup(ctx context.Context, in *ReleaseRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/fcdaemon.FcService/ReleaseVMSetup", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fcServiceClient) CreateCPUHogs(ctx context.Context, in *CpuHogs, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/fcdaemon.FcService/CreateCPUHogs", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *fcServiceClient) StopCPUHogs(ctx context.Context, in *CpuHogs, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/fcdaemon.FcService/StopCPUHogs", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// FcServiceServer is the server API for FcService service.
type FcServiceServer interface {
	// Asks the fcdaemon to create ONE replica
	CreateReplica(context.Context, *CreateRequest) (*CreateReply, error)
	ReleaseVMSetup(context.Context, *ReleaseRequest) (*emptypb.Empty, error)
	CreateCPUHogs(context.Context, *CpuHogs) (*emptypb.Empty, error)
	StopCPUHogs(context.Context, *CpuHogs) (*emptypb.Empty, error)
}

// UnimplementedFcServiceServer can be embedded to have forward compatible implementations.
type UnimplementedFcServiceServer struct {
}

func (*UnimplementedFcServiceServer) CreateReplica(context.Context, *CreateRequest) (*CreateReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateReplica not implemented")
}
func (*UnimplementedFcServiceServer) ReleaseVMSetup(context.Context, *ReleaseRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReleaseVMSetup not implemented")
}
func (*UnimplementedFcServiceServer) CreateCPUHogs(context.Context, *CpuHogs) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateCPUHogs not implemented")
}
func (*UnimplementedFcServiceServer) StopCPUHogs(context.Context, *CpuHogs) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method StopCPUHogs not implemented")
}

func RegisterFcServiceServer(s *grpc.Server, srv FcServiceServer) {
	s.RegisterService(&_FcService_serviceDesc, srv)
}

func _FcService_CreateReplica_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FcServiceServer).CreateReplica(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/fcdaemon.FcService/CreateReplica",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FcServiceServer).CreateReplica(ctx, req.(*CreateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FcService_ReleaseVMSetup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReleaseRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FcServiceServer).ReleaseVMSetup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/fcdaemon.FcService/ReleaseVMSetup",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FcServiceServer).ReleaseVMSetup(ctx, req.(*ReleaseRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _FcService_CreateCPUHogs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CpuHogs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FcServiceServer).CreateCPUHogs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/fcdaemon.FcService/CreateCPUHogs",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FcServiceServer).CreateCPUHogs(ctx, req.(*CpuHogs))
	}
	return interceptor(ctx, in, info, handler)
}

func _FcService_StopCPUHogs_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CpuHogs)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(FcServiceServer).StopCPUHogs(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/fcdaemon.FcService/StopCPUHogs",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(FcServiceServer).StopCPUHogs(ctx, req.(*CpuHogs))
	}
	return interceptor(ctx, in, info, handler)
}

var _FcService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "fcdaemon.FcService",
	HandlerType: (*FcServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CreateReplica",
			Handler:    _FcService_CreateReplica_Handler,
		},
		{
			MethodName: "ReleaseVMSetup",
			Handler:    _FcService_ReleaseVMSetup_Handler,
		},
		{
			MethodName: "CreateCPUHogs",
			Handler:    _FcService_CreateCPUHogs_Handler,
		},
		{
			MethodName: "StopCPUHogs",
			Handler:    _FcService_StopCPUHogs_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "fc_daemon_service.proto",
}
