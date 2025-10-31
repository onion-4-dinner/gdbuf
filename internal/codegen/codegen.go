package codegen

import (
	"embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/huandu/xstrings"
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
	FileName             string
	ProtoPath            string // original path of the proto file in the proto module that was injested
	ProtoPathNoExtension string // original path of the proto file in the proto module that was injested
	ClassName            string // camel case without .proto suffix
	Messages             []protoMessage
}

type protoMessage struct {
	MessageName string
	Fields      []protoMessageField
}

type protoMessageField struct {
	FieldName string
	GodotType string
}

func NewCodeGenerator(logger *slog.Logger, destinationDirectoryPath, extensionName, protobufVersion string) (*CodeGenerator, error) {
	tmpl, err := template.ParseFS(templatesFS, "templates/**/*.tmpl", "templates/**/**/*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("could not parse code generation file templates: %w", err)
	}

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
	protoData, err := extractProtoData(fileDescriptorSet)
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
		cg.logger.Info("processing file", "name", file.FileName)
		thisProtoFileTemplates := map[string]string{
			"refcounted.h.tmpl":   fmt.Sprintf("src/%s.h", strings.TrimSuffix(file.FileName, ".proto")),
			"refcounted.cpp.tmpl": fmt.Sprintf("src/%s.cpp", strings.TrimSuffix(file.FileName, ".proto")),
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

func extractProtoData(fileDescriptorSet []*descriptorpb.FileDescriptorProto) (*protoData, error) {
	var protoData protoData
	for _, file := range fileDescriptorSet {
		var protoFile protoFile
		protoFile.ProtoPath = file.GetName()
		protoFile.ProtoPathNoExtension = strings.TrimSuffix(protoFile.ProtoPath, ".proto")
		protoFile.FileName = filepath.Base(protoFile.ProtoPathNoExtension)
		protoFile.ClassName = xstrings.ToPascalCase(protoFile.FileName)
		for _, msg := range file.GetMessageType() {
			var protoMessage protoMessage
			protoMessage.MessageName = msg.GetName()
			for _, field := range msg.GetField() {
				var protoMessageField protoMessageField
				protoMessageField.FieldName = field.GetName()
				godotType, err := mapProtoToGodotType(*field.GetType().Enum())
				if err != nil {
					return nil, fmt.Errorf("could not map proto to godot type: %w", err)
				}
				protoMessageField.GodotType = godotType
				protoMessage.Fields = append(protoMessage.Fields, protoMessageField)
			}
			protoFile.Messages = append(protoFile.Messages, protoMessage)
		}
		protoData.Files = append(protoData.Files, protoFile)
	}
	return &protoData, nil
}
