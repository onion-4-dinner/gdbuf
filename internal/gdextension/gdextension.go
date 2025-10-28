package gdextension

import (
	"embed"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
)

//go:embed build/*
var buildEnvFS embed.FS

type GDExtensionBuilder struct {
	logger *slog.Logger
}

func NewGDExtensionBuilder(logger *slog.Logger) *GDExtensionBuilder {
	return &GDExtensionBuilder{
		logger: logger,
	}
}

func (gde *GDExtensionBuilder) Build(customBuildFilesDir string) error {
	buildDir, err := os.MkdirTemp("", "gdbuf-build-")
	if err != nil {
		return fmt.Errorf("could not make build directory: %w", err)
	}

	gde.logger.Info("starting gdextension build", "build_dir", buildDir)

	buildEnv, err := fs.Sub(buildEnvFS, "build")
	if err != nil {
		return fmt.Errorf("could not make build directory: %w", err)
	}

	if err = os.CopyFS(buildDir, buildEnv); err != nil {
		return fmt.Errorf("could not copy build environment to temp directory: %w", err)
	}

	if err = os.CopyFS(buildDir, os.DirFS(customBuildFilesDir)); err != nil {
		return fmt.Errorf("could not copy custom build files to temp directory: %w", err)
	}

	// all files are in place, try to build
	buildCmd := exec.Command("make", "build-linux")
	buildCmd.Dir = buildDir
	output, err := buildCmd.CombinedOutput()
	if err != nil {
		gde.logger.Error("build error", "command", buildCmd.Args, "output", output)
		return fmt.Errorf("build error: %w", err)
	}
	gde.logger.Info("build successful")
	return nil
}
