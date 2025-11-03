package codegen

import (
	"fmt"

	"google.golang.org/protobuf/types/descriptorpb"
)

var protoGodotTypeMap = map[descriptorpb.FieldDescriptorProto_Type]string{
	descriptorpb.FieldDescriptorProto_TYPE_STRING:   "godot::String",
	descriptorpb.FieldDescriptorProto_TYPE_BOOL:     "bool",
	descriptorpb.FieldDescriptorProto_TYPE_INT32:    "int32_t",
	descriptorpb.FieldDescriptorProto_TYPE_INT64:    "int64_t",
	descriptorpb.FieldDescriptorProto_TYPE_FLOAT:    "float",
	descriptorpb.FieldDescriptorProto_TYPE_BYTES:    "godot::PackedByteArray",
	descriptorpb.FieldDescriptorProto_TYPE_DOUBLE:   "double",
	descriptorpb.FieldDescriptorProto_TYPE_UINT32:   "uint32_t",
	descriptorpb.FieldDescriptorProto_TYPE_UINT64:   "uint64_t",
	descriptorpb.FieldDescriptorProto_TYPE_SINT32:   "int32_t",
	descriptorpb.FieldDescriptorProto_TYPE_SINT64:   "int64_t",
	descriptorpb.FieldDescriptorProto_TYPE_FIXED32:  "int32_t",
	descriptorpb.FieldDescriptorProto_TYPE_FIXED64:  "int64_t",
	descriptorpb.FieldDescriptorProto_TYPE_SFIXED32: "int32_t",
	descriptorpb.FieldDescriptorProto_TYPE_SFIXED64: "int64_t",
	// descriptorpb.FieldDescriptorProto_TYPE_MESSAGE: "UNKNOWN",
	// descriptorpb.FieldDescriptorProto_TYPE_ENUM:    "UNKNOWN",
}

func mapProtoToGodotType(protoType descriptorpb.FieldDescriptorProto_Type) (string, error) {
	godotType, ok := protoGodotTypeMap[protoType]
	if !ok {
		return "", fmt.Errorf("unknown or unsupported proto type: %s", protoType.String())
	}

	return godotType, nil
}
