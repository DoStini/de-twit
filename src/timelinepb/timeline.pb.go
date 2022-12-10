// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.9
// source: protobuf/timeline.proto

package timelinepb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Timeline struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// TODO: USER
	Posts []*Post `protobuf:"bytes,1,rep,name=posts,proto3" json:"posts,omitempty"`
}

func (x *Timeline) Reset() {
	*x = Timeline{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protobuf_timeline_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Timeline) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Timeline) ProtoMessage() {}

func (x *Timeline) ProtoReflect() protoreflect.Message {
	mi := &file_protobuf_timeline_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Timeline.ProtoReflect.Descriptor instead.
func (*Timeline) Descriptor() ([]byte, []int) {
	return file_protobuf_timeline_proto_rawDescGZIP(), []int{0}
}

func (x *Timeline) GetPosts() []*Post {
	if x != nil {
		return x.Posts
	}
	return nil
}

type Post struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Id          string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Text        string                 `protobuf:"bytes,2,opt,name=text,proto3" json:"text,omitempty"`
	User        string                 `protobuf:"bytes,3,opt,name=user,proto3" json:"user,omitempty"`
	LastUpdated *timestamppb.Timestamp `protobuf:"bytes,4,opt,name=last_updated,json=lastUpdated,proto3" json:"last_updated,omitempty"`
}

func (x *Post) Reset() {
	*x = Post{}
	if protoimpl.UnsafeEnabled {
		mi := &file_protobuf_timeline_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Post) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Post) ProtoMessage() {}

func (x *Post) ProtoReflect() protoreflect.Message {
	mi := &file_protobuf_timeline_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Post.ProtoReflect.Descriptor instead.
func (*Post) Descriptor() ([]byte, []int) {
	return file_protobuf_timeline_proto_rawDescGZIP(), []int{1}
}

func (x *Post) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Post) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

func (x *Post) GetUser() string {
	if x != nil {
		return x.User
	}
	return ""
}

func (x *Post) GetLastUpdated() *timestamppb.Timestamp {
	if x != nil {
		return x.LastUpdated
	}
	return nil
}

var File_protobuf_timeline_proto protoreflect.FileDescriptor

var file_protobuf_timeline_proto_rawDesc = []byte{
	0x0a, 0x17, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x6c,
	0x69, 0x6e, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x74, 0x69, 0x6d, 0x65, 0x6c,
	0x69, 0x6e, 0x65, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x22, 0x30, 0x0a, 0x08, 0x54, 0x69, 0x6d, 0x65, 0x6c, 0x69, 0x6e, 0x65,
	0x12, 0x24, 0x0a, 0x05, 0x70, 0x6f, 0x73, 0x74, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32,
	0x0e, 0x2e, 0x74, 0x69, 0x6d, 0x65, 0x6c, 0x69, 0x6e, 0x65, 0x2e, 0x50, 0x6f, 0x73, 0x74, 0x52,
	0x05, 0x70, 0x6f, 0x73, 0x74, 0x73, 0x22, 0x7d, 0x0a, 0x04, 0x50, 0x6f, 0x73, 0x74, 0x12, 0x0e,
	0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12,
	0x0a, 0x04, 0x74, 0x65, 0x78, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x65,
	0x78, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x75, 0x73, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x04, 0x75, 0x73, 0x65, 0x72, 0x12, 0x3d, 0x0a, 0x0c, 0x6c, 0x61, 0x73, 0x74, 0x5f, 0x75,
	0x70, 0x64, 0x61, 0x74, 0x65, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54,
	0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x0b, 0x6c, 0x61, 0x73, 0x74, 0x55, 0x70,
	0x64, 0x61, 0x74, 0x65, 0x64, 0x42, 0x10, 0x5a, 0x0e, 0x73, 0x72, 0x63, 0x2f, 0x74, 0x69, 0x6d,
	0x65, 0x6c, 0x69, 0x6e, 0x65, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_protobuf_timeline_proto_rawDescOnce sync.Once
	file_protobuf_timeline_proto_rawDescData = file_protobuf_timeline_proto_rawDesc
)

func file_protobuf_timeline_proto_rawDescGZIP() []byte {
	file_protobuf_timeline_proto_rawDescOnce.Do(func() {
		file_protobuf_timeline_proto_rawDescData = protoimpl.X.CompressGZIP(file_protobuf_timeline_proto_rawDescData)
	})
	return file_protobuf_timeline_proto_rawDescData
}

var file_protobuf_timeline_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_protobuf_timeline_proto_goTypes = []interface{}{
	(*Timeline)(nil),              // 0: timeline.Timeline
	(*Post)(nil),                  // 1: timeline.Post
	(*timestamppb.Timestamp)(nil), // 2: google.protobuf.Timestamp
}
var file_protobuf_timeline_proto_depIdxs = []int32{
	1, // 0: timeline.Timeline.posts:type_name -> timeline.Post
	2, // 1: timeline.Post.last_updated:type_name -> google.protobuf.Timestamp
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_protobuf_timeline_proto_init() }
func file_protobuf_timeline_proto_init() {
	if File_protobuf_timeline_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_protobuf_timeline_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Timeline); i {
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
		file_protobuf_timeline_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Post); i {
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
			RawDescriptor: file_protobuf_timeline_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_protobuf_timeline_proto_goTypes,
		DependencyIndexes: file_protobuf_timeline_proto_depIdxs,
		MessageInfos:      file_protobuf_timeline_proto_msgTypes,
	}.Build()
	File_protobuf_timeline_proto = out.File
	file_protobuf_timeline_proto_rawDesc = nil
	file_protobuf_timeline_proto_goTypes = nil
	file_protobuf_timeline_proto_depIdxs = nil
}
