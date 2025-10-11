package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/LJ-Software/gdbuf/internal/codegen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))
	logger.Info("starting gdbuf")

	protoDescFilePathPtr := flag.String("proto-desc", "", "path to generated proto description file")
	genOutDirPathPtr := flag.String("out", ".", "generated proto c++ code output path")

	flag.Parse()

	if len(*protoDescFilePathPtr) == 0 {
		logger.Error("required argument --proto-desc not given")
		os.Exit(1)
	}

	if err := checkPath(*protoDescFilePathPtr, false); err != nil {
		logger.Error("invalid path for proto description file", "err", err)
		os.Exit(1)
	}

	if err := checkPath(*genOutDirPathPtr, true); err != nil {
		logger.Error("invalid path for output directory", "err", err)
		os.Exit(1)
	}

	protoDescData, err := os.ReadFile(*protoDescFilePathPtr)
	if err != nil {
		logger.Error("could not read proto description file", "err", err)
		os.Exit(1)
	}

	var protoFileDescriptorSet descriptorpb.FileDescriptorSet
	if err = proto.Unmarshal(protoDescData, &protoFileDescriptorSet); err != nil {
		logger.Error("could not unmarshal proto description data", "err", err)
		os.Exit(1)
	}

	codeGenerator, err := codegen.NewCodeGenerator(logger, *genOutDirPathPtr)
	if err != nil {
		logger.Error("could not create new code generator", "err", err)
		os.Exit(1)
	}

	err = codeGenerator.GenerateCode(protoFileDescriptorSet.File)
	if err != nil {
		logger.Error("problem generating code", "err", err)
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
