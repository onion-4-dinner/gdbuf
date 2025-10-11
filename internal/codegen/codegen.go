package codegen

import (
	"embed"
	"fmt"
	"log/slog"
	"path/filepath"
	"text/template"

	"google.golang.org/protobuf/types/descriptorpb"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

type CodeGenerator struct {
	logger                   *slog.Logger
	destinationDirectoryPath string
	templates                *template.Template
}

func NewCodeGenerator(logger *slog.Logger, destinationDirectoryPath string) (*CodeGenerator, error) {
	tmpl, err := template.ParseFS(templatesFS, "templates/*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("could not parse code generation file templates: %w", err)
	}

	templates := tmpl.Templates()
	for _, t := range templates {
		logger.Info("template loaded", "name", t.Name())
	}

	return &CodeGenerator{
		logger:                   logger,
		destinationDirectoryPath: destinationDirectoryPath,
		templates:                tmpl,
	}, nil
}

func (cg *CodeGenerator) GenerateCode(fileDescriptorSet []*descriptorpb.FileDescriptorProto) error {
	for _, protoFileDescriptor := range fileDescriptorSet {
		fileName := protoFileDescriptor.GetName()
		cg.logger.Info("processing file", "name", fileName)
		if err := cg.generateCppFiles(protoFileDescriptor); err != nil {
			return fmt.Errorf("problem generating cpp file at %s: %w", filepath.Join(cg.destinationDirectoryPath, fileName), err)
		}
	}

	return nil
}

func (cg *CodeGenerator) generateCppFiles(protoFileDescriptor *descriptorpb.FileDescriptorProto) error {
	cppTemplateData, err := newCppTemplateData(protoFileDescriptor)
	if err != nil {
		return fmt.Errorf("could not parse cpp template data: %w", err)
	}

	err = cg.executeTemplate("refcounted.h.tmpl", filepath.Join(cg.destinationDirectoryPath, protoFileDescriptor.GetName()+".h"), cppTemplateData)
	if err != nil {
		return fmt.Errorf("could not execute template: %w", err)
	}

	err = cg.executeTemplate("refcounted.cpp.tmpl", filepath.Join(cg.destinationDirectoryPath, protoFileDescriptor.GetName()+".cpp"), cppTemplateData)
	if err != nil {
		return fmt.Errorf("could not execute template: %w", err)
	}

	return nil
}
