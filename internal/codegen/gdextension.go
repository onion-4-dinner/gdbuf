package codegen

type gdextensionTemplateData struct {
	ExtensionName   string
	ProtobufVersion string
}

func newGdextensionTemplateData(gdextensionName, protobufVersion string) (*gdextensionTemplateData, error) {
	return &gdextensionTemplateData{
		ExtensionName:   gdextensionName,
		ProtobufVersion: protobufVersion,
	}, nil
}
