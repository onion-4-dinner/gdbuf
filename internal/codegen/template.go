package codegen

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/huandu/xstrings"
	"google.golang.org/protobuf/types/descriptorpb"
)

type cppFileTemplateData struct {
	FileName string
	Messages []Message
}

type Message struct {
	Name   string
	Fields []MessageField
}

type MessageField struct {
	Name      string
	GodotType string
}

func newCppTemplateData(protoFileDescriptor *descriptorpb.FileDescriptorProto) (*cppFileTemplateData, error) {
	var cppTemplateData cppFileTemplateData

	cppTemplateData.FileName = xstrings.ToPascalCase(filepath.Base(protoFileDescriptor.GetName()))

	for _, message := range protoFileDescriptor.GetMessageType() {
		fields := []MessageField{}
		for _, field := range message.GetField() {
			godotType, err := mapProtoToGodotType(field.GetType())
			if err != nil {
				return nil, fmt.Errorf("could not map proto type to godot type: %w", err)
			}
			fields = append(fields, MessageField{
				Name:      field.GetName(),
				GodotType: godotType,
			})
		}
		cppTemplateData.Messages = append(cppTemplateData.Messages, Message{
			Name:   message.GetName(),
			Fields: fields,
		})
	}

	return &cppTemplateData, nil
}

func (cg *CodeGenerator) executeTemplate(templateName, outFilePath string, templateData any) error {

	err := os.MkdirAll(filepath.Dir(outFilePath), 0777)
	if err != nil {
		return fmt.Errorf("could not create directory at %s: %w", filepath.Dir(outFilePath), err)
	}

	outFile, err := os.Create(outFilePath)
	if err != nil {
		return fmt.Errorf("could not create file at %s: %w", outFilePath, err)
	}

	err = cg.templates.ExecuteTemplate(outFile, templateName, templateData)
	if err != nil {
		return fmt.Errorf("could not execute template %s: %w", templateName, err)
	}

	return nil
}
