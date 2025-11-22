package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/LJ-Software/gdbuf/internal/codegen"
	"github.com/LJ-Software/gdbuf/internal/gdextension"
	"github.com/LJ-Software/gdbuf/internal/protoc"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	logger.Info("starting gdbuf")

	protoInputDirPtr := flag.String("proto", "", "path to proto definition files")
	cppOutputDirPtr := flag.String("genout", ".", "generated proto c++ code output path")
	extensionNamePtr := flag.String("name", "gdbufgen", "name of the generated gdextension")
	extensionArtifactOutputDirPtr := flag.String("out", "./out", "output directory location of the generated gdextension")

	flag.Parse()

	if len(*protoInputDirPtr) == 0 {
		logger.Error("required argument --proto not given")
		os.Exit(1)
	}

	if err := checkPath(*protoInputDirPtr, true); err != nil {
		logger.Error("invalid path for proto files", "err", err)
		os.Exit(1)
	}

	if err := checkPath(*cppOutputDirPtr, true); err != nil {
		logger.Error("invalid path for code gen output directory", "err", err)
		os.Exit(1)
	}

	if err := checkPath(*extensionArtifactOutputDirPtr, true); err != nil {
		logger.Error("invalid path for gdextension output directory", "err", err)
		os.Exit(1)
	}

	protoc, err := protoc.NewProtoCompiler(logger.WithGroup("protoc"))
	if err != nil {
		logger.Error("could not create new proto compiler", "err", err)
		os.Exit(1)
	}

	descriptorSet, err := protoc.BuildDescriptorSet(*protoInputDirPtr)
	if err != nil {
		logger.Error("could not build descriptor set for protobuf definitions", "err", err)
		os.Exit(1)
	}

	compiledProtoCppTempDirPath, err := protoc.CompileCpp(*protoInputDirPtr)
	if err != nil {
		logger.Error("could not compile proto cpp", "err", err)
		os.Exit(1)
	}

	codeGenerator, err := codegen.NewCodeGenerator(logger, *cppOutputDirPtr, *extensionNamePtr, protoc.GetVersion())
	if err != nil {
		logger.Error("could not create new code generator", "err", err)
		os.Exit(1)
	}

	err = codeGenerator.GenerateCode(descriptorSet)
	if err != nil {
		logger.Error("problem generating code", "err", err)
		os.Exit(1)
	}

	compiledProtoCppOutDirPath := filepath.Join(*cppOutputDirPtr, "src")
	err = os.CopyFS(compiledProtoCppOutDirPath, os.DirFS(compiledProtoCppTempDirPath))
	if err != nil {
		logger.Error("problem copying compiled cpp proto to directory", "err", err)
		os.Exit(1)
	}

	gdExtensionBuilder := gdextension.NewGDExtensionBuilder(logger)

	err = gdExtensionBuilder.Build(*cppOutputDirPtr, *extensionArtifactOutputDirPtr)
	if err != nil {
		logger.Error("problem building gdextension", "err", err)
		os.Exit(1)
	}
}

func checkPath(path string, isDir bool) error {
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		return errors.New("path does not exist")
	} else if err != nil {
		return fmt.Errorf("problem with path: %w", err)
	} else if fileInfo.IsDir() != isDir {
		return fmt.Errorf("isDir expected: %t but got: %t", isDir, fileInfo.IsDir())
	}
	return nil
}
