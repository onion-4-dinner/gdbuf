package codegen

import (
	"fmt"
	"os"
	"path/filepath"
)

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
