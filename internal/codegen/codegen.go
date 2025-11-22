package codegen

import (
	"embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
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
	ProtoPath   string // original path of the proto file in the proto module that was injested
	PackageName string
	Messages    []protoMessage
	Enums       []protoEnum
}

type protoMessage struct {
	MessageName string
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
}

func NewCodeGenerator(logger *slog.Logger, destinationDirectoryPath, extensionName, protobufVersion string) (*CodeGenerator, error) {
	tmpl, err := template.New("gdbuf").Funcs(sprig.FuncMap()).ParseFS(templatesFS, "templates/**/*.tmpl", "templates/**/**/*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("could not parse code generation file templates: %w", err)
	}

	tmpl = tmpl.Funcs(sprig.FuncMap())
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

		for _, msg := range file.GetMessageType() {
			var protoMessage protoMessage
			protoMessage.MessageName = msg.GetName()

			for _, field := range msg.GetField() {
				var protoMessageField protoMessageField
				protoMessageField.FieldName = field.GetName()
				protoMessageField.ProtoTypeName = field.GetTypeName()

				godotType, isCustom, isEnum, err := resolveGodotType(field, protoFile.ProtoPath, protoFileToDeclaredMessageNames, protoFileToDeclaredEnumNames)
				if err != nil {
					return nil, fmt.Errorf("could not resolve godot type: %w", err)
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
