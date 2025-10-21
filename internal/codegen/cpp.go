package codegen

import (
	"fmt"
	"path/filepath"
	"strings"

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

	cppTemplateData.FileName = xstrings.ToPascalCase(strings.TrimSuffix(filepath.Base(protoFileDescriptor.GetName()), ".proto"))

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
