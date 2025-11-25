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

type godotTypeInfo struct {
	GodotType      string
	GodotClassName string
}

var wktMap = map[string]godotTypeInfo{
	"Timestamp":   {"int64_t", "int"},
	"Duration":    {"double", "float"},
	"Struct":      {"godot::Dictionary", "Dictionary"},
	"Any":         {"godot::Dictionary", "Dictionary"},
	"ListValue":   {"godot::Array", "Array"},
	"Value":       {"godot::Variant", "Variant"},
	"Empty":       {"godot::Variant", "Variant"},
	"StringValue": {"godot::String", "String"},
	"Int32Value":  {"int32_t", "int"},
	"BoolValue":   {"bool", "bool"},
	"FieldMask":   {"godot::PackedStringArray", "PackedStringArray"},
}

func resolveGodotType(field *descriptorpb.FieldDescriptorProto, currentProtoPath string, fileToMsgs map[string][]string, fileToEnum map[string][]string, allMessageDescriptors map[string]*descriptorpb.DescriptorProto, typeToGodotName map[string]string) (godotType string, godotClassName string, isCustom bool, isEnum bool, srcFile string, err error) {
	fieldType := *field.GetType().Enum()
	fullTypeName := field.GetTypeName()

	srcFile = ""

	switch fieldType {
	case descriptorpb.FieldDescriptorProto_TYPE_MESSAGE:
		if strings.HasPrefix(fullTypeName, ".google.protobuf.") {
			srcFile = "google::protobuf"
		} else {
			// Check if it's a MapEntry
			if desc, ok := allMessageDescriptors[fullTypeName]; ok {
				if desc.GetOptions().GetMapEntry() {
					return "godot::Dictionary", "Dictionary", false, false, "", nil
				}
			}

			for f := range fileToMsgs {
				if slices.Contains(fileToMsgs[f], fullTypeName) {
					srcFile = f
					break
				}
			}
		}

		if srcFile == "" {
			return "", "", false, false, "", fmt.Errorf("could not find source file for message type: %s", fullTypeName)
		}

		if srcFile == "google::protobuf" {
			isCustom = false
			// Handle WKTs by short name
			shortName := fullTypeName
			if lastDot := strings.LastIndex(fullTypeName, "."); lastDot != -1 {
				shortName = fullTypeName[lastDot+1:]
			}

			if info, ok := wktMap[shortName]; ok {
				godotType = info.GodotType
				godotClassName = info.GodotClassName
			} else {
				isCustom = true
				godotType = fmt.Sprintf("google::protobuf::%s", shortName)
				godotClassName = shortName
			}
		} else {
			isCustom = true
			var ok bool
			godotClassName, ok = typeToGodotName[fullTypeName]
			if !ok {
				// Fallback or error? Should be there.
				// Try simple name
				if lastDot := strings.LastIndex(fullTypeName, "."); lastDot != -1 {
					godotClassName = fullTypeName[lastDot+1:]
				} else {
					godotClassName = fullTypeName
				}
			}
			godotClassName = toPascalCase(godotClassName)

			if srcFile == currentProtoPath {
				godotType = godotClassName
			} else {
				godotType = fmt.Sprintf("gdbuf::%s::%s", filepath.Base(strings.TrimSuffix(srcFile, ".proto")), godotClassName)
			}
		}
	case descriptorpb.FieldDescriptorProto_TYPE_ENUM:
		for f := range fileToEnum {
			if slices.Contains(fileToEnum[f], fullTypeName) {
				srcFile = f
				break
			}
		}
		if srcFile == "" {
			return "", "", false, false, "", fmt.Errorf("could not find source file for enum type: %s", fullTypeName)
		}
		isCustom = false
		isEnum = true
		godotType = "int32_t" // Bind as int to avoid Godot binding issues with C++ enums
		godotClassName = "int"
	default:
		isCustom = false
		var ok bool
		godotType, ok = protoGodotTypeMap[fieldType]
		if !ok {
			return "", "", false, false, "", fmt.Errorf("unknown or unsupported proto type: %s", fieldType.String())
		}
		godotClassName = godotType // Simplification, cleaned up later if needed, but usually not used for primitives in hints
	}

	return godotType, godotClassName, isCustom, isEnum, srcFile, nil
}
