package codegen

import (
	"embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"text/template"

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

func NewCodeGenerator(logger *slog.Logger, destinationDirectoryPath, extensionName, protobufVersion string) (*CodeGenerator, error) {
	tmpl, err := template.ParseFS(templatesFS, "templates/**/*.tmpl")
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
	// TODO: gen the one time files before the per-proto files
	// generate once per gdextension
	if err := cg.generateGdextensionBuildFiles(cg.extensionName, cg.protobufVersion); err != nil {
		return fmt.Errorf("problem generating gdextension build files at %s: %w", cg.destinationDirectoryPath, err)
	}
	if err := cg.generateProtoCppOneTimeFiles(); err != nil {
		return fmt.Errorf("problem generating gdextension build files at %s: %w", cg.destinationDirectoryPath, err)
	}

	// generate for each proto file
	for _, protoFileDescriptor := range fileDescriptorSet {
		fileName := protoFileDescriptor.GetName()
		cg.logger.Info("processing file", "name", fileName)
		if err := cg.generateProtoCppFile(protoFileDescriptor); err != nil {
			return fmt.Errorf("problem generating cpp file at %s: %w", filepath.Join(cg.destinationDirectoryPath, fileName), err)
		}
	}

	return nil
}

func (cg *CodeGenerator) generateProtoCppFile(protoFileDescriptor *descriptorpb.FileDescriptorProto) error {
	err := os.MkdirAll(filepath.Join(cg.destinationDirectoryPath, "src"), 0755)
	if err != nil {
		return fmt.Errorf("could not make cpp output directory: %w", err)
	}

	cppTemplateData, err := newCppTemplateData(protoFileDescriptor)
	if err != nil {
		return fmt.Errorf("could not parse cpp template data: %w", err)
	}

	err = cg.executeTemplate("register_types.h.tmpl", filepath.Join(cg.destinationDirectoryPath, "src", "register_types.h"), cppTemplateData)
	if err != nil {
		return fmt.Errorf("could not execute template: %w", err)
	}

	err = cg.executeTemplate("register_types.cpp.tmpl", filepath.Join(cg.destinationDirectoryPath, "src", "register_types.cpp"), cppTemplateData)
	if err != nil {
		return fmt.Errorf("could not execute template: %w", err)
	}

	err = cg.executeTemplate("refcounted.h.tmpl", filepath.Join(cg.destinationDirectoryPath, "src", protoFileDescriptor.GetName()+".h"), cppTemplateData)
	if err != nil {
		return fmt.Errorf("could not execute template: %w", err)
	}

	err = cg.executeTemplate("refcounted.cpp.tmpl", filepath.Join(cg.destinationDirectoryPath, "src", protoFileDescriptor.GetName()+".cpp"), cppTemplateData)
	if err != nil {
		return fmt.Errorf("could not execute template: %w", err)
	}

	return nil
}

func (cg *CodeGenerator) generateGdextensionBuildFiles(extensionName, protobufVersion string) error {
	err := os.MkdirAll(filepath.Join(cg.destinationDirectoryPath, "out", extensionName, "dist"), 0755)
	if err != nil {
		return fmt.Errorf("could not make cpp output directory: %w", err)
	}

	gdextensionTemplateData, err := newGdextensionTemplateData(extensionName, protobufVersion)
	if err != nil {
		return fmt.Errorf("could not parse cpp template data: %w", err)
	}

	err = cg.executeTemplate("CMakeLists.txt.tmpl", filepath.Join(cg.destinationDirectoryPath, "CMakeLists.txt"), gdextensionTemplateData)
	if err != nil {
		return fmt.Errorf("could not execute template: %w", err)
	}

	err = cg.executeTemplate("vcpkg.json.tmpl", filepath.Join(cg.destinationDirectoryPath, "vcpkg.json"), gdextensionTemplateData)
	if err != nil {
		return fmt.Errorf("could not execute template: %w", err)
	}

	err = cg.executeTemplate("gde-protobuf.gdextension.tmpl", filepath.Join(cg.destinationDirectoryPath, "out", extensionName, fmt.Sprintf("%s.gdextension", extensionName)), gdextensionTemplateData)
	if err != nil {
		return fmt.Errorf("could not execute template: %w", err)
	}

	return nil
}

func (cg *CodeGenerator) generateProtoCppOneTimeFiles() error {
	err := os.MkdirAll(filepath.Join(cg.destinationDirectoryPath, "src"), 0755)
	if err != nil {
		return fmt.Errorf("could not make cpp output directory: %w", err)
	}

	protoCppOneTimeTemplateData, err := newProtoCppOneTimeTemplateData([]string{})
	if err != nil {
		return fmt.Errorf("could not parse cpp template data: %w", err)
	}

	err = cg.executeTemplate("register_types.h.tmpl", filepath.Join(cg.destinationDirectoryPath, "src", "register_types.h"), protoCppOneTimeTemplateData)
	if err != nil {
		return fmt.Errorf("could not execute template: %w", err)
	}
	err = cg.executeTemplate("register_types.cpp.tmpl", filepath.Join(cg.destinationDirectoryPath, "src", "register_types.cpp"), protoCppOneTimeTemplateData)
	if err != nil {
		return fmt.Errorf("could not execute template: %w", err)
	}

	return nil
}
