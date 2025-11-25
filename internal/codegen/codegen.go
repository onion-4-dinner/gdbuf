package codegen

import (
	"embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

//go:embed templates/*
var templatesFS embed.FS

func toPascalCase(s string) string {
	parts := strings.Split(s, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}

type CodeGenerator struct {
	logger                   *slog.Logger
	destinationDirectoryPath string
	templates                *template.Template
	extensionName            string
	protobufVersion          string
}

type templateData struct {
	GDExtensionName string
	ProtobufVersion string
	ProtoData       protoData
}

type protoData struct {
	Files []protoFile
}

type protoFile struct {
	ProtoPath    string // original path of the proto file in the proto module that was injested
	PackageName  string
	Messages     []protoMessage
	Enums        []protoEnum
	Dependencies []string
	ForwardDecls []ForwardDecl
}

type ForwardDecl struct {
	Namespace string
	ClassName string
}

type protoMessage struct {
	ClassName   string
	MessageName string
	Description string
	Fields      []protoMessageField
	Oneofs      []protoOneof
}

type protoOneof struct {
	Name   string
	Fields []protoMessageField
}

type protoOneofField struct {
	Name   string
	Number int32
}

type protoEnum struct {
	EnumName string
	Options  []string
}

type protoMessageField struct {
	FieldName           string
	ProtoTypeName       string
	GodotType           string
	GodotClassName      string
	InnerGodotType      string
	InnerGodotClassName string
	IsCustomType        bool
	IsInnerCustomType   bool
	IsRepeated          bool
	IsEnum              bool
	IsMap               bool
	MapKeyGodotType     string
	MapValueGodotType   string
	MapValueIsCustom    bool
	Description         string
	OneofName           string
	Number              int32
}

func NewCodeGenerator(logger *slog.Logger, destinationDirectoryPath, extensionName, protobufVersion string) (*CodeGenerator, error) {
	tmpl, err := template.New("gdbuf").Funcs(getTemplateFuncMap()).ParseFS(templatesFS, "templates/**/*.tmpl", "templates/**/**/*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("could not parse code generation file templates: %w", err)
	}

	tmpl = tmpl.Funcs(getTemplateFuncMap())
	templates := tmpl.Templates()
	for _, t := range templates {
		logger.Debug("template loaded", "name", t.Name())
	}

	return &CodeGenerator{
		logger:                   logger,
		destinationDirectoryPath: destinationDirectoryPath,
		templates:                tmpl,
		extensionName:            extensionName,
		protobufVersion:          protobufVersion,
	}, nil
}

func getTemplateFuncMap() template.FuncMap {
	f := sprig.FuncMap()
	f["toPascalCase"] = toPascalCase
	f["toUpper"] = strings.ToUpper
	f["nanopbType"] = func(protoType string) string {
		// Remove leading dot
		s := strings.TrimPrefix(protoType, ".")
		// Replace dots with underscores
		return strings.ReplaceAll(s, ".", "_")
	}
	f["godotVariantType"] = func(godotType string, isCustom bool, isEnum bool) string {
		if isEnum {
			return "godot::Variant::INT"
		}
		if isCustom {
			return "godot::Variant::OBJECT"
		}
		switch godotType {
		case "bool":
			return "godot::Variant::BOOL"
		case "int32_t", "int64_t", "uint32_t", "uint64_t":
			return "godot::Variant::INT"
		case "float", "double":
			return "godot::Variant::FLOAT"
		case "godot::String":
			return "godot::Variant::STRING"
		case "godot::PackedByteArray":
			return "godot::Variant::PACKED_BYTE_ARRAY"
		case "godot::PackedStringArray":
			return "godot::Variant::PACKED_STRING_ARRAY"
		case "godot::Dictionary":
			return "godot::Variant::DICTIONARY"
		case "godot::Array":
			return "godot::Variant::ARRAY"
		default:
			return "godot::Variant::NIL"
		}
	}
	f["godotDocType"] = func(godotType string, isCustom bool, isEnum bool) string {
		if isEnum {
			return "int"
		}
		if isCustom {
			parts := strings.Split(godotType, "::")
			return parts[len(parts)-1]
		}
		switch godotType {
		case "bool":
			return "bool"
		case "int32_t", "int64_t", "uint32_t", "uint64_t":
			return "int"
		case "float", "double":
			return "float"
		case "godot::String":
			return "String"
		case "godot::PackedByteArray":
			return "PackedByteArray"
		case "godot::PackedStringArray":
			return "PackedStringArray"
		case "godot::Dictionary":
			return "Dictionary"
		case "godot::Array":
			return "Array"
		default:
			return "Variant"
		}
	}
	return f
}

func (cg *CodeGenerator) GenerateCode(fileDescriptorSet []*descriptorpb.FileDescriptorProto) error {
	// first extract all of the data needed
	protoData, err := cg.extractProtoData(fileDescriptorSet)
	if err != nil {
		return fmt.Errorf("problem extracting proto data: %w", err)
	}

	templateData := templateData{
		GDExtensionName: cg.extensionName,
		ProtobufVersion: cg.protobufVersion,
		ProtoData:       *protoData,
	}

	oneTimeTemplates := map[string]string{
		"CMakeLists.txt.tmpl":           "CMakeLists.txt",
		"gde-protobuf.gdextension.tmpl": "out/gde-protobuf.gdextension",
		"register_types.h.tmpl":         "src/register_types.h",
		"register_types.cpp.tmpl":       "src/register_types.cpp",
		"messages.h.tmpl":               "src/messages.h",
		"messages.cpp.tmpl":             "src/messages.cpp",
	}

	for templateName, outputPath := range oneTimeTemplates {
		err := os.MkdirAll(filepath.Dir(filepath.Join(cg.destinationDirectoryPath, outputPath)), 0755)
		if err != nil {
			return fmt.Errorf("could not make template output directory: %w", err)
		}
		if err := cg.executeTemplate(templateName, filepath.Join(cg.destinationDirectoryPath, outputPath), templateData); err != nil {
			return fmt.Errorf("could not execute template %s: %w", templateName, err)
		}
	}

	// generate for each proto file
	for _, file := range templateData.ProtoData.Files {
		cg.logger.Info("processing file", "name", file.ProtoPath)
		thisProtoFileTemplates := map[string]string{
			"resource.h.tmpl":   fmt.Sprintf("src/%s.h", strings.TrimSuffix(file.ProtoPath, ".proto")),
			"resource.cpp.tmpl": fmt.Sprintf("src/%s.cpp", strings.TrimSuffix(file.ProtoPath, ".proto")),
		}
		for templateName, outputPath := range thisProtoFileTemplates {
			err := os.MkdirAll(filepath.Dir(filepath.Join(cg.destinationDirectoryPath, outputPath)), 0755)
			if err != nil {
				return fmt.Errorf("could not make template output directory: %w", err)
			}
			if err := cg.executeTemplate(templateName, filepath.Join(cg.destinationDirectoryPath, outputPath), file); err != nil {
				return fmt.Errorf("could not execute template %s: %w", templateName, err)
			}
		}

		for _, msg := range file.Messages {
			className := msg.ClassName
			outputPath := filepath.Join(cg.destinationDirectoryPath, "doc_classes", className+".xml")
			if err := cg.executeTemplate("class_doc.xml.tmpl", outputPath, msg); err != nil {
				return fmt.Errorf("could not execute template class_doc.xml.tmpl for message %s: %w", msg.MessageName, err)
			}
		}
	}

	return nil
}

func (cg *CodeGenerator) extractProtoData(fileDescriptorSet []*descriptorpb.FileDescriptorProto) (*protoData, error) {
	var protoData protoData
	// one loop through to get a mapping of filename and message name (recursive)
	var protoFileToDeclaredMessageNames map[string][]string = make(map[string][]string)
	var protoFileToDeclaredEnumNames map[string][]string = make(map[string][]string)
	var allMessageDescriptors map[string]*descriptorpb.DescriptorProto = make(map[string]*descriptorpb.DescriptorProto)
	var typeToGodotName map[string]string = make(map[string]string)

	for _, file := range fileDescriptorSet {
		pkg := file.GetPackage()
		prefix := "."
		if pkg != "" {
			prefix = "." + pkg + "."
		}

		// Recursive message traversal
		var traverseMsgs func(msgs []*descriptorpb.DescriptorProto, currentPrefix string)
		traverseMsgs = func(msgs []*descriptorpb.DescriptorProto, currentPrefix string) {
			for _, msg := range msgs {
				fullName := currentPrefix + msg.GetName()
				protoFileToDeclaredMessageNames[file.GetName()] = append(protoFileToDeclaredMessageNames[file.GetName()], fullName)
				allMessageDescriptors[fullName] = msg

				// Construct Godot Name
				// Remove package prefix
				shortName := strings.TrimPrefix(fullName, prefix)
				// Replace . with _
				godotName := strings.ReplaceAll(shortName, ".", "_")
				typeToGodotName[fullName] = godotName

				traverseMsgs(msg.GetNestedType(), fullName+".")

				// Also traverse Nested Enums
				for _, enum := range msg.GetEnumType() {
					enumFullName := fullName + "." + enum.GetName()
					protoFileToDeclaredEnumNames[file.GetName()] = append(protoFileToDeclaredEnumNames[file.GetName()], enumFullName)
				}
			}
		}
		traverseMsgs(file.GetMessageType(), prefix)

		// Top level enums
		for _, enum := range file.GetEnumType() {
			fullName := prefix + enum.GetName()
			protoFileToDeclaredEnumNames[file.GetName()] = append(protoFileToDeclaredEnumNames[file.GetName()], fullName)
		}
	}

	// now for the real deal
	for _, file := range fileDescriptorSet {
		var protoFile protoFile
		protoFile.ProtoPath = file.GetName()
		protoFile.PackageName = file.GetPackage()
		cg.logger.Info("processing proto file", "name", file.GetName(), "package", protoFile.PackageName)

		pkg := file.GetPackage()
		prefix := "."
		if pkg != "" {
			prefix = "." + pkg + "."
		}

		for _, enum := range file.GetEnumType() {
			var protoEnum protoEnum
			protoEnum.EnumName = enum.GetName()
			enum.GetOptions().ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
				protoEnum.Options = append(protoEnum.Options, fd.TextName())
				return true
			})
			protoFile.Enums = append(protoFile.Enums, protoEnum)
		}

		// Recursive generation
		var messagesToGenerate []protoMessage
		var traverseGen func(msgs []*descriptorpb.DescriptorProto, currentPrefix string, path []int32) error
		traverseGen = func(msgs []*descriptorpb.DescriptorProto, currentPrefix string, path []int32) error {
			for msgIndex, msg := range msgs {
				if msg.GetOptions().GetMapEntry() {
					continue
				}

				fullName := currentPrefix + msg.GetName()
				godotName := typeToGodotName[fullName]

				var protoMessage protoMessage
				protoMessage.MessageName = godotName
				protoMessage.ClassName = toPascalCase(godotName)
				currentPath := append(slices.Clone(path), int32(msgIndex))
				protoMessage.Description = getComments(file.GetSourceCodeInfo(), currentPath)

				// Process Oneofs
				oneofDecls := msg.GetOneofDecl()
				for i, oneof := range oneofDecls {
					protoMessage.Oneofs = append(protoMessage.Oneofs, protoOneof{
						Name: oneof.GetName(),
					})
					// Note: We will populate Fields later as we iterate fields
					// Actually, we might want to verify if it's synthetic.
					// But proto3 optional uses oneofs too.
					// We can check if any field in this oneof is optional.
					// Ideally we just map them.
					_ = i
				}

				for fieldIndex, field := range msg.GetField() {
					var protoMessageField protoMessageField
					protoMessageField.FieldName = field.GetName()
					protoMessageField.ProtoTypeName = field.GetTypeName()
					protoMessageField.Number = field.GetNumber()
					fieldPath := append(slices.Clone(currentPath), 2, int32(fieldIndex))
					protoMessageField.Description = getComments(file.GetSourceCodeInfo(), fieldPath)

					if field.OneofIndex != nil {
						// Check if it is NOT a synthetic proto3 optional
						if !field.GetProto3Optional() {
							oneofIdx := field.GetOneofIndex()
							if int(oneofIdx) < len(protoMessage.Oneofs) {
								protoMessageField.OneofName = protoMessage.Oneofs[oneofIdx].Name
							}
						}
					}

					// Add default descriptions for WKTs
					if protoMessageField.ProtoTypeName == ".google.protobuf.Timestamp" {
						if protoMessageField.Description != "" {
							protoMessageField.Description += "\n"
						}
						protoMessageField.Description += "Note: This field is a Google Protobuf Timestamp. In Godot, it is represented as an int64 (Unix timestamp in milliseconds)."
					} else if protoMessageField.ProtoTypeName == ".google.protobuf.Duration" {
						if protoMessageField.Description != "" {
							protoMessageField.Description += "\n"
						}
						protoMessageField.Description += "Note: This field is a Google Protobuf Duration. In Godot, it is represented as a double (seconds)."
					} else if protoMessageField.ProtoTypeName == ".google.protobuf.Struct" {
						if protoMessageField.Description != "" {
							protoMessageField.Description += "\n"
						}
						protoMessageField.Description += "Note: This field is a Google Protobuf Struct. In Godot, it is represented as a Dictionary."
					}

					godotType, godotClassName, isCustom, isEnum, srcFile, err := resolveGodotType(field, protoFile.ProtoPath, protoFileToDeclaredMessageNames, protoFileToDeclaredEnumNames, allMessageDescriptors, typeToGodotName)
					if err != nil {
						return fmt.Errorf("could not resolve godot type: %w", err)
					}

					if isCustom && srcFile != "" && srcFile != "google::protobuf" && srcFile != protoFile.ProtoPath {
						headerPath := strings.TrimSuffix(srcFile, ".proto") + ".h"
						if !slices.Contains(protoFile.Dependencies, headerPath) {
							protoFile.Dependencies = append(protoFile.Dependencies, headerPath)
						}

						// Add forward declaration
						parts := strings.Split(godotType, "::")
						if len(parts) > 1 {
							className := parts[len(parts)-1]
							namespace := strings.Join(parts[:len(parts)-1], "::")
							fd := ForwardDecl{Namespace: namespace, ClassName: className}
							exists := false
							for _, existing := range protoFile.ForwardDecls {
								if existing == fd {
									exists = true
									break
								}
							}
							if !exists {
								protoFile.ForwardDecls = append(protoFile.ForwardDecls, fd)
							}
						}
					}

					protoMessageField.IsCustomType = isCustom
					protoMessageField.IsEnum = isEnum

					protoMessageField.InnerGodotType = godotType
					protoMessageField.InnerGodotClassName = godotClassName
					protoMessageField.IsInnerCustomType = protoMessageField.IsCustomType

					if desc, ok := allMessageDescriptors[field.GetTypeName()]; ok && desc.GetOptions().GetMapEntry() {
						protoMessageField.IsMap = true
						var keyField, valueField *descriptorpb.FieldDescriptorProto
						for _, f := range desc.GetField() {
							if f.GetNumber() == 1 {
								keyField = f
							}
							if f.GetNumber() == 2 {
								valueField = f
							}
						}
						keyType, _, _, _, _, err := resolveGodotType(keyField, protoFile.ProtoPath, protoFileToDeclaredMessageNames, protoFileToDeclaredEnumNames, allMessageDescriptors, typeToGodotName)
						if err != nil {
							return fmt.Errorf("could not resolve map key type: %w", err)
						}
						valType, _, valCustom, _, _, err := resolveGodotType(valueField, protoFile.ProtoPath, protoFileToDeclaredMessageNames, protoFileToDeclaredEnumNames, allMessageDescriptors, typeToGodotName)
						if err != nil {
							return fmt.Errorf("could not resolve map value type: %w", err)
						}
						protoMessageField.MapKeyGodotType = keyType
						protoMessageField.MapValueGodotType = valType
						protoMessageField.MapValueIsCustom = valCustom
						protoMessageField.GodotType = "godot::Dictionary"
						protoMessageField.GodotClassName = "Dictionary"
						// For maps, we might need to know the value's class name if it's a custom object?
						// The template for maps uses .MapValueGodotType (C++ type).
						// If MapValueIsCustom is true, it does cast_to<{{ .MapValueGodotType }}>.
						// It doesn't seem to use property hints for the dictionary content explicitly in the template I read earlier?
						// Wait, let's recheck resource.cpp.tmpl.
						protoMessageField.IsRepeated = false
					} else {
						protoMessageField.GodotType = godotType
						protoMessageField.GodotClassName = godotClassName
						if field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED {
							cg.logger.Debug("Repeated field", "name", field.GetName(), "before", protoMessageField.IsCustomType)
							protoMessageField.IsRepeated = true
							protoMessageField.GodotType = "godot::Array"
							protoMessageField.GodotClassName = "Array"
							protoMessageField.IsCustomType = false
							cg.logger.Debug("Repeated field", "name", field.GetName(), "after", protoMessageField.IsCustomType)
						}
					}

					if field.OneofIndex != nil && !field.GetProto3Optional() {
						oneofIdx := field.GetOneofIndex()
						if int(oneofIdx) < len(protoMessage.Oneofs) {
							protoMessage.Oneofs[oneofIdx].Fields = append(protoMessage.Oneofs[oneofIdx].Fields, protoMessageField)
						}
					}

					protoMessage.Fields = append(protoMessage.Fields, protoMessageField)
				}
				cg.logger.Debug("Generated message", "name", protoMessage.MessageName, "fields", len(protoMessage.Fields))
				messagesToGenerate = append(messagesToGenerate, protoMessage)

				// Recurse for nested types. NestedType is field 3.
				nestedPath := append(slices.Clone(currentPath), 3)
				if err := traverseGen(msg.GetNestedType(), fullName+".", nestedPath); err != nil {
					return err
				}
			}
			return nil
		}

		// MessageType is field 4 in FileDescriptorProto
		if err := traverseGen(file.GetMessageType(), prefix, []int32{4}); err != nil {
			return nil, err
		}
		protoFile.Messages = messagesToGenerate

		protoData.Files = append(protoData.Files, protoFile)
	}
	return &protoData, nil
}

func getComments(sc *descriptorpb.SourceCodeInfo, path []int32) string {
	if sc == nil {
		return ""
	}
	for _, loc := range sc.GetLocation() {
		if slices.Equal(loc.Path, path) {
			c := strings.TrimSpace(loc.GetLeadingComments())
			if c == "" {
				c = strings.TrimSpace(loc.GetTrailingComments())
			}
			// Clean up comments: Remove leading * or // if present (protoc usually handles this but sometimes...)
			// Actually GetLeadingComments usually returns the raw comment content.
			return strings.TrimSpace(c)
		}
	}
	return ""
}
