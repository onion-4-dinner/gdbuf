package codegen

import (
	"testing"

	"google.golang.org/protobuf/types/descriptorpb"
)

func TestResolveGodotType(t *testing.T) {
	// Helper to create string pointer
	s := func(v string) *string { return &v }
	// Helper to create type enum pointer
	typeEnum := func(v descriptorpb.FieldDescriptorProto_Type) *descriptorpb.FieldDescriptorProto_Type { return &v }

	tests := []struct {
		name             string
		field            *descriptorpb.FieldDescriptorProto
		currentProtoPath string
		fileToMsgs       map[string][]string
		fileToEnum       map[string][]string
		wantType         string
		wantIsCustom     bool
		wantIsEnum       bool
		wantErr          bool
	}{
		{
			name: "Primitive Int32",
			field: &descriptorpb.FieldDescriptorProto{
				Type:     typeEnum(descriptorpb.FieldDescriptorProto_TYPE_INT32),
				TypeName: s("int32"),
			},
			wantType:     "int32_t",
			wantIsCustom: false,
		},
		{
			name: "WKT Timestamp",
			field: &descriptorpb.FieldDescriptorProto{
				Type:     typeEnum(descriptorpb.FieldDescriptorProto_TYPE_MESSAGE),
				TypeName: s(".google.protobuf.Timestamp"),
			},
			wantType:     "int64_t",
			wantIsCustom: false,
		},
		{
			name: "WKT Struct",
			field: &descriptorpb.FieldDescriptorProto{
				Type:     typeEnum(descriptorpb.FieldDescriptorProto_TYPE_MESSAGE),
				TypeName: s(".google.protobuf.Struct"),
			},
			wantType:     "godot::Dictionary",
			wantIsCustom: false,
		},
		{
			name: "Custom Message Same File",
			field: &descriptorpb.FieldDescriptorProto{
				Type:     typeEnum(descriptorpb.FieldDescriptorProto_TYPE_MESSAGE),
				TypeName: s(".MyMessage"),
			},
			currentProtoPath: "test.proto",
			fileToMsgs: map[string][]string{
				"test.proto": {"MyMessage"},
			},
			wantType:     "MyMessage",
			wantIsCustom: true,
		},
		{
			name: "Custom Message Other File",
			field: &descriptorpb.FieldDescriptorProto{
				Type:     typeEnum(descriptorpb.FieldDescriptorProto_TYPE_MESSAGE),
				TypeName: s(".OtherMessage"),
			},
			currentProtoPath: "test.proto",
			fileToMsgs: map[string][]string{
				"other.proto": {"OtherMessage"},
			},
			wantType:     "other::OtherMessage",
			wantIsCustom: true,
		},
		{
			name: "Enum",
			field: &descriptorpb.FieldDescriptorProto{
				Type:     typeEnum(descriptorpb.FieldDescriptorProto_TYPE_ENUM),
				TypeName: s(".MyEnum"),
			},
			fileToEnum: map[string][]string{
				"test.proto": {"MyEnum"},
			},
			wantType:   "int32_t",
			wantIsEnum: true,
		},
		{
			name: "Unknown Message Type",
			field: &descriptorpb.FieldDescriptorProto{
				Type:     typeEnum(descriptorpb.FieldDescriptorProto_TYPE_MESSAGE),
				TypeName: s(".UnknownMessage"),
			},
			fileToMsgs: map[string][]string{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotIsCustom, gotIsEnum, _, err := resolveGodotType(tt.field, tt.currentProtoPath, tt.fileToMsgs, tt.fileToEnum, nil, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveGodotType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if gotType != tt.wantType {
					t.Errorf("resolveGodotType() gotType = %v, want %v", gotType, tt.wantType)
				}
				if gotIsCustom != tt.wantIsCustom {
					t.Errorf("resolveGodotType() gotIsCustom = %v, want %v", gotIsCustom, tt.wantIsCustom)
				}
				if gotIsEnum != tt.wantIsEnum {
					t.Errorf("resolveGodotType() gotIsEnum = %v, want %v", gotIsEnum, tt.wantIsEnum)
				}
			}
		})
	}
}
