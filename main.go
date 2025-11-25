package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/LJ-Software/gdbuf/internal/codegen"
	"github.com/LJ-Software/gdbuf/internal/gdextension"
	"github.com/LJ-Software/gdbuf/internal/protoc"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, ",")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	logger.Info("starting gdbuf")

	var includeDirs arrayFlags
	flag.Var(&includeDirs, "include", "include directories for proto files")
	protoInputDirPtr := flag.String("proto", "", "path to proto definition files")
	cppOutputDirPtr := flag.String("genout", ".", "generated proto c++ code output path")
	extensionNamePtr := flag.String("name", "gdbufgen", "name of the generated gdextension")
	extensionArtifactOutputDirPtr := flag.String("out", "./out", "output directory location of the generated gdextension")
	generateOnlyPtr := flag.Bool("generate-only", false, "only generate c++ code, do not compile gdextension")

	flag.Parse()

	if len(*protoInputDirPtr) == 0 {
		logger.Error("required argument --proto not given")
		os.Exit(1)
	}

	if err := checkPath(*protoInputDirPtr, true); err != nil {
		logger.Error("invalid path for proto files", "err", err)
		os.Exit(1)
	}

	for _, includeDir := range includeDirs {
		if err := checkPath(includeDir, true); err != nil {
			logger.Error("invalid path for include directory", "dir", includeDir, "err", err)
			os.Exit(1)
		}
	}

	if err := checkPath(*cppOutputDirPtr, true); err != nil {
		logger.Error("invalid path for code gen output directory", "err", err)
		os.Exit(1)
	}

	if err := checkPath(*extensionArtifactOutputDirPtr, true); err != nil {
		logger.Error("invalid path for gdextension output directory", "err", err)
		os.Exit(1)
	}

	// Prepare Nanopb Generator (extracted to temp)
	genTmpDir, err := os.MkdirTemp("", "nanopb-gen-")
	if err != nil {
		logger.Error("could not create temp dir for generator", "err", err)
		os.Exit(1)
	}
	defer os.RemoveAll(genTmpDir)

	gdExtensionBuilder := gdextension.NewGDExtensionBuilder(logger)
	if err := gdExtensionBuilder.ExtractNanopbGenerator(genTmpDir); err != nil {
		logger.Error("could not extract nanopb generator", "err", err)
		os.Exit(1)
	}

	protoc, err := protoc.NewProtoCompiler(logger.WithGroup("protoc"))
	if err != nil {
		logger.Error("could not create new proto compiler", "err", err)
		os.Exit(1)
	}

	descriptorSet, err := protoc.BuildDescriptorSet(*protoInputDirPtr, includeDirs)
	if err != nil {
		logger.Error("could not build descriptor set for protobuf definitions", "err", err)
		os.Exit(1)
	}

	compiledProtoCppTempDirPath, err := protoc.CompileNanopb(*protoInputDirPtr, includeDirs, genTmpDir)
	if err != nil {
		logger.Error("could not compile proto cpp (nanopb)", "err", err)
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
	err = copyDir(compiledProtoCppTempDirPath, compiledProtoCppOutDirPath)
	if err != nil {
		logger.Error("problem copying compiled cpp proto to directory", "err", err)
		os.Exit(1)
	}

	err = gdExtensionBuilder.Build(*cppOutputDirPtr, *extensionArtifactOutputDirPtr, *generateOnlyPtr)
	if err != nil {
		logger.Error("problem building gdextension", "err", err)
		os.Exit(1)
	}
}

func copyDir(src string, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
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
