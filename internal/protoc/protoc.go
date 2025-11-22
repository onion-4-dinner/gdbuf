package protoc

import (
	"bytes"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

type ProtoCompiler struct {
	logger          *slog.Logger
	protobufVersion string
}

func NewProtoCompiler(logger *slog.Logger) (*ProtoCompiler, error) {
	version, err := getProtocExecutableVersion()
	if err != nil {
		return nil, fmt.Errorf("could not get protoc executable version information: %w", err)
	}
	logger.Debug("got protobuf version", "version", version)
	return &ProtoCompiler{
		logger:          logger,
		protobufVersion: version,
	}, nil
}

func (c *ProtoCompiler) GetVersion() string {
	return c.protobufVersion
}

func (c *ProtoCompiler) BuildDescriptorSet(protoFilesDirPath string) ([]*descriptorpb.FileDescriptorProto, error) {
	var descriptorSet []*descriptorpb.FileDescriptorProto

	protoFilePaths, err := getProtoFilesInDir(protoFilesDirPath)
	if err != nil {
		return descriptorSet, fmt.Errorf("could not get proto files from %s: %w", protoFilesDirPath, err)
	}

	tmpDir := os.TempDir()
	protoDescriptorPath := filepath.Join(tmpDir, "gdbuf.desc.binpb")
	buildProtoDescriptorCmd := exec.Command("protoc", append([]string{fmt.Sprintf("--descriptor_set_out=%s", protoDescriptorPath)}, protoFilePaths...)...)

	var stderr bytes.Buffer
	buildProtoDescriptorCmd.Stderr = &stderr

	err = buildProtoDescriptorCmd.Run()
	if err != nil {
		return descriptorSet, fmt.Errorf("could not build proto description file with cmd [%v]: %s", buildProtoDescriptorCmd.Args, stderr.String())
	}
	c.logger.Info("generated protobuf descriptor file", "path", protoDescriptorPath)

	protoDescData, err := os.ReadFile(protoDescriptorPath)
	if err != nil {
		return descriptorSet, fmt.Errorf("could not read proto description file: %w", err)
	}

	var protoFileDescriptorSet descriptorpb.FileDescriptorSet
	if err = proto.Unmarshal(protoDescData, &protoFileDescriptorSet); err != nil {
		return descriptorSet, fmt.Errorf("could not unmarshal proto description data: %w", err)
	}

	return protoFileDescriptorSet.GetFile(), nil
}

func (c *ProtoCompiler) CompileCpp(protoFilesDirPath string) (string, error) {
	tempProtocBuildDir, err := os.MkdirTemp("", "gdbuf-build-")
	if err != nil {
		return "", fmt.Errorf("could not make temp directory for proto cpp build: %w", err)
	}

	protoFilePaths, err := getProtoFilesInDir(protoFilesDirPath)
	if err != nil {
		return "", fmt.Errorf("could not get proto files from %s: %w", protoFilesDirPath, err)
	}
	compileCppCmd := exec.Command("protoc", append([]string{fmt.Sprintf("--cpp_out=%s", tempProtocBuildDir)}, protoFilePaths...)...)

	var stderr bytes.Buffer
	compileCppCmd.Stderr = &stderr

	err = compileCppCmd.Run()
	if err != nil {
		return "", fmt.Errorf("could not compile cpp proto files with cmd [%v]: %s", compileCppCmd.Args, stderr.String())
	}

	return tempProtocBuildDir, nil
}

func getProtocExecutableVersion() (string, error) {
	protoVersionCmdOut, err := exec.Command("protoc", "--version").Output()
	if err != nil {
		return "", fmt.Errorf("could not build proto description file: %w", err)
	}

	// Output format: "libprotoc 25.1" or similar
	re := regexp.MustCompile(`libprotoc (\d+\.\d+)`)
	matches := re.FindStringSubmatch(string(protoVersionCmdOut))
	if len(matches) < 2 {
		return "", fmt.Errorf("could not parse protoc version from output: %s", string(protoVersionCmdOut))
	}

	versionParts := strings.Split(matches[1], ".")
	protoVersion := fmt.Sprintf("%s.%s.5", versionParts[1], versionParts[0]) // TODO: dont hardcode last part of version
	return protoVersion, nil
}

func getProtoFilesInDir(protoFilesDirPath string) ([]string, error) {
	protoFilePaths := []string{}
	err := fs.WalkDir(os.DirFS(protoFilesDirPath), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("dir walk err: %w", err)
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), ".proto") {
			protoFilePaths = append(protoFilePaths, filepath.Join(protoFilesDirPath, path))
		}
		return nil
	})
	if err != nil {
		return protoFilePaths, fmt.Errorf("could not walk proto file directory: %w", err)
	}
	return protoFilePaths, nil
}
