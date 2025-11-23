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
	"github.com/huandu/xstrings"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

//go:embed templates/*
var templatesFS embed.FS

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
	MessageName string
	Description string
	Fields      []protoMessageField
}

type protoEnum struct {
	EnumName string
	Options  []string
}

type protoMessageField struct {
	FieldName         string
	ProtoTypeName     string
	GodotType         string
	InnerGodotType    string
	IsCustomType      bool
	IsInnerCustomType bool
	IsRepeated        bool
	IsEnum            bool
	Description       string
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
		case "godot::Dictionary":
			return "godot::Variant::DICTIONARY"
		case "godot::Array":
			return "godot::Variant::ARRAY"
		default:
			return "godot::Variant::NIL"
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
		"vcpkg.json.tmpl":               "vcpkg.json",
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
			"refcounted.h.tmpl":   fmt.Sprintf("src/%s.h", strings.TrimSuffix(file.ProtoPath, ".proto")),
			"refcounted.cpp.tmpl": fmt.Sprintf("src/%s.cpp", strings.TrimSuffix(file.ProtoPath, ".proto")),
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
			className := xstrings.ToCamelCase(msg.MessageName)
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
	// one loop through to get a mapping of filename and message name
	var protoFileToDeclaredMessageNames map[string][]string = make(map[string][]string)
	var protoFileToDeclaredEnumNames map[string][]string = make(map[string][]string)
	for _, file := range fileDescriptorSet {
		for _, msg := range file.GetMessageType() {
			cg.logger.Debug("found message", "file", file.GetName(), "message", msg.GetName())
			protoFileToDeclaredMessageNames[file.GetName()] = append(protoFileToDeclaredMessageNames[file.GetName()], msg.GetName())
		}
		for _, enum := range file.GetEnumType() {
			cg.logger.Debug("found enum", "file", file.GetName(), "enum", enum.GetName())
			protoFileToDeclaredEnumNames[file.GetName()] = append(protoFileToDeclaredEnumNames[file.GetName()], enum.GetName())
		}
	}

	// now for the real deal
	for _, file := range fileDescriptorSet {
		var protoFile protoFile
		protoFile.ProtoPath = file.GetName()
		protoFile.PackageName = file.GetPackage()

		for _, enum := range file.GetEnumType() {
			var protoEnum protoEnum
			protoEnum.EnumName = enum.GetName()
			enum.GetOptions().ProtoReflect().Range(func(fd protoreflect.FieldDescriptor, v protoreflect.Value) bool {
				protoEnum.Options = append(protoEnum.Options, fd.TextName())
				return true
			})
			protoFile.Enums = append(protoFile.Enums, protoEnum)
		}

		for msgIndex, msg := range file.GetMessageType() {
			var protoMessage protoMessage
			protoMessage.MessageName = msg.GetName()
			protoMessage.Description = getComments(file.GetSourceCodeInfo(), []int32{4, int32(msgIndex)})

			for fieldIndex, field := range msg.GetField() {
				var protoMessageField protoMessageField
				protoMessageField.FieldName = field.GetName()
				protoMessageField.ProtoTypeName = field.GetTypeName()
				protoMessageField.Description = getComments(file.GetSourceCodeInfo(), []int32{4, int32(msgIndex), 2, int32(fieldIndex)})

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

				godotType, isCustom, isEnum, srcFile, err := resolveGodotType(field, protoFile.ProtoPath, protoFileToDeclaredMessageNames, protoFileToDeclaredEnumNames)
				if err != nil {
					return nil, fmt.Errorf("could not resolve godot type: %w", err)
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

				protoMessageField.GodotType = godotType
				protoMessageField.IsCustomType = isCustom
				protoMessageField.IsEnum = isEnum

				protoMessageField.InnerGodotType = godotType
				protoMessageField.IsInnerCustomType = protoMessageField.IsCustomType

				if field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED {
					protoMessageField.IsRepeated = true
					protoMessageField.GodotType = "godot::Array"
					protoMessageField.IsCustomType = false
				} else {
					protoMessageField.GodotType = godotType
				}

				protoMessage.Fields = append(protoMessage.Fields, protoMessageField)
			}
			protoFile.Messages = append(protoFile.Messages, protoMessage)
		}
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
