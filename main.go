package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/LJ-Software/gdbuf/internal/codegen"
	"github.com/LJ-Software/gdbuf/internal/protoc"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	logger.Info("starting gdbuf")

	protoFilePathPtr := flag.String("proto", "", "path to proto definition files")
	genOutDirPathPtr := flag.String("out", ".", "generated proto c++ code output path")
	extensionNamePtr := flag.String("name", "gdbufgen", "name of the generated gdextension")

	flag.Parse()

	if len(*protoFilePathPtr) == 0 {
		logger.Error("required argument --proto not given")
		os.Exit(1)
	}

	if err := checkPath(*protoFilePathPtr, true); err != nil {
		logger.Error("invalid path for proto files", "err", err)
		os.Exit(1)
	}

	if err := checkPath(*genOutDirPathPtr, true); err != nil {
		logger.Error("invalid path for output directory", "err", err)
		os.Exit(1)
	}

	protoc, err := protoc.NewProtoCompiler(logger.WithGroup("protoc"))
	if err != nil {
		logger.Error("could not create new proto compiler", "err", err)
		os.Exit(1)
	}

	descriptorSet, err := protoc.BuildDescriptorSet(*protoFilePathPtr)
	if err != nil {
		logger.Error("could not build descriptor set for protobuf definitions", "err", err)
		os.Exit(1)
	}

	compiledProtoCppTempDirPath, err := protoc.CompileCpp(*protoFilePathPtr)
	if err != nil {
		logger.Error("could not compile proto cpp", "err", err)
		os.Exit(1)
	}

	codeGenerator, err := codegen.NewCodeGenerator(logger, *genOutDirPathPtr, *extensionNamePtr, protoc.GetVersion())
	if err != nil {
		logger.Error("could not create new code generator", "err", err)
		os.Exit(1)
	}

	err = codeGenerator.GenerateCode(descriptorSet)
	if err != nil {
		logger.Error("problem generating code", "err", err)
		os.Exit(1)
	}

	compiledProtoCppOutDirPath := filepath.Join(*genOutDirPathPtr, "src", "proto")
	err = os.CopyFS(compiledProtoCppOutDirPath, os.DirFS(compiledProtoCppTempDirPath))
	if err != nil {
		logger.Error("problem copying compiled cpp proto to directory", "err", err)
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
