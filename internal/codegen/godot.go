package codegen

import (
	"fmt"

	"google.golang.org/protobuf/types/descriptorpb"
)

var protoGodotTypeMap = map[descriptorpb.FieldDescriptorProto_Type]string{
	descriptorpb.FieldDescriptorProto_TYPE_STRING: "Variant::STRING",
	descriptorpb.FieldDescriptorProto_TYPE_BOOL:   "Variant::BOOL",
	descriptorpb.FieldDescriptorProto_TYPE_INT32:  "Variant::INT",
	descriptorpb.FieldDescriptorProto_TYPE_INT64:  "Variant::INT",
	descriptorpb.FieldDescriptorProto_TYPE_FLOAT:  "Variant::FLOAT",
	descriptorpb.FieldDescriptorProto_TYPE_BYTES:  "Variant::PACKED_BYTE_ARRAY",
	descriptorpb.FieldDescriptorProto_TYPE_DOUBLE: "Variant::FLOAT",
	// TODO: each of these need a mapping
	descriptorpb.FieldDescriptorProto_TYPE_UINT32:   "UNKNOWN",
	descriptorpb.FieldDescriptorProto_TYPE_UINT64:   "UNKNOWN",
	descriptorpb.FieldDescriptorProto_TYPE_SINT32:   "UNKNOWN",
	descriptorpb.FieldDescriptorProto_TYPE_SINT64:   "UNKNOWN",
	descriptorpb.FieldDescriptorProto_TYPE_FIXED32:  "UNKNOWN",
	descriptorpb.FieldDescriptorProto_TYPE_FIXED64:  "UNKNOWN",
	descriptorpb.FieldDescriptorProto_TYPE_SFIXED32: "UNKNOWN",
	descriptorpb.FieldDescriptorProto_TYPE_SFIXED64: "UNKNOWN",
	descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:  "UNKNOWN",
	descriptorpb.FieldDescriptorProto_TYPE_ENUM:     "UNKNOWN",
}

func mapProtoToGodotType(protoType descriptorpb.FieldDescriptorProto_Type) (string, error) {
	godotType, ok := protoGodotTypeMap[protoType]
	if !ok {
		return "", fmt.Errorf("unknown or unsupported proto type: %s", protoType.String())
	}

	return godotType, nil
}
