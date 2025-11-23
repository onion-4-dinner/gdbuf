package codegen

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

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
}

func resolveGodotType(field *descriptorpb.FieldDescriptorProto, currentProtoPath string, fileToMsgs map[string][]string, fileToEnum map[string][]string) (godotType string, isCustom bool, isEnum bool, srcFile string, err error) {
	fieldType := *field.GetType().Enum()
	// Get the plain type name (e.g. "Timestamp" from ".google.protobuf.Timestamp")
	fieldTypeName := field.GetTypeName()
	if lastDot := strings.LastIndex(fieldTypeName, "."); lastDot != -1 {
		fieldTypeName = fieldTypeName[lastDot+1:]
	}

	srcFile = ""

	switch fieldType {
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		switch fieldTypeName {
		case "Any", "Timestamp", "Duration", "StringValue", "Int32Value", "Struct", "ListValue", "Value", "Empty", "BoolValue":
			srcFile = "google::protobuf"
		default:
			for f := range fileToMsgs {
				if slices.Contains(fileToMsgs[f], fieldTypeName) {
					srcFile = f
					break
				}
			}
		}
		if srcFile == "" {
			return "", false, false, "", fmt.Errorf("could not find source file for message type: %s", fieldTypeName)
		}

		if srcFile == "google::protobuf" {
			isCustom = false
			switch fieldTypeName {
			case "Timestamp":
				godotType = "int64_t"
			case "Duration":
				godotType = "double"
			case "Struct", "Any":
				godotType = "godot::Dictionary"
			case "ListValue":
				godotType = "godot::Array"
			case "Value", "Empty":
				godotType = "godot::Variant"
			case "StringValue":
				godotType = "godot::String"
			case "Int32Value":
				godotType = "int32_t"
			case "BoolValue":
				godotType = "bool"
			default:
				isCustom = true
				godotType = fmt.Sprintf("google::protobuf::%s", fieldTypeName)
			}
		} else {
			isCustom = true
			if srcFile == currentProtoPath {
				godotType = fieldTypeName
			} else {
				godotType = fmt.Sprintf("gdbuf::%s::%s", filepath.Base(strings.TrimSuffix(srcFile, ".proto")), fieldTypeName)
			}
		}
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		for f := range fileToEnum {
			if slices.Contains(fileToEnum[f], fieldTypeName) {
				srcFile = f
				break
			}
		}
		if srcFile == "" {
			return "", false, false, "", fmt.Errorf("could not find source file for enum type: %s", fieldTypeName)
		}
		isCustom = false
		isEnum = true
		godotType = "int32_t" // Bind as int to avoid Godot binding issues with C++ enums
	default:
		isCustom = false
		var ok bool
		godotType, ok = protoGodotTypeMap[fieldType]
		if !ok {
			return "", false, false, "", fmt.Errorf("unknown or unsupported proto type: %s", fieldType.String())
		}
	}

	return godotType, isCustom, isEnum, srcFile, nil
}
