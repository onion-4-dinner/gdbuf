package gdextension

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

//go:embed all:buildenv/*
var buildEnvFS embed.FS

type GDExtensionBuilder struct {
	logger *slog.Logger
}

func NewGDExtensionBuilder(logger *slog.Logger) *GDExtensionBuilder {
	return &GDExtensionBuilder{
		logger: logger,
	}
}

func (gde *GDExtensionBuilder) Build(generatedCppSourceDir, outputDir string) error {
	buildDir, err := os.MkdirTemp("", "gdbuf-build-")
	if err != nil {
		return fmt.Errorf("could not make build directory: %w", err)
	}

	gde.logger.Info("starting gdextension build", "build_dir", buildDir)

	buildEnv, err := fs.Sub(buildEnvFS, "buildenv")
	if err != nil {
		return fmt.Errorf("could not make build directory: %w", err)
	}

	if err = copyFS(buildEnv, buildDir); err != nil {
		return fmt.Errorf("could not copy build environment to temp directory: %w", err)
	}

	if err = copyFS(os.DirFS(generatedCppSourceDir), buildDir); err != nil {
		return fmt.Errorf("could not copy custom build files to temp directory: %w", err)
	}

	// all files are in place, try to build
	var buildTarget string
	switch runtime.GOOS {
	case "linux":
		buildTarget = "build-linux"
	case "darwin":
		buildTarget = "build-macos"
	case "windows":
		buildTarget = "build-windows"
	default:
		return fmt.Errorf("unsupported os: %s", runtime.GOOS)
	}

	buildCmd := exec.Command("make", buildTarget)
	buildCmd.Env = os.Environ()
	buildCmd.Env = append(buildCmd.Env, fmt.Sprintf("VCPKG_ROOT=%s", filepath.Join(buildDir, "vcpkg")))
	buildCmd.Env = append(buildCmd.Env, fmt.Sprintf("WORKSPACE=%s", buildDir))
	buildCmd.Dir = buildDir
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	err = buildCmd.Run()
	if err != nil {
		return fmt.Errorf("build error: %w", err)
	}
	gde.logger.Info("build successful")

	if err = copyFS(os.DirFS(filepath.Join(buildDir, "build", "bin")), filepath.Join(buildDir, "out", "dist")); err != nil {
		return fmt.Errorf("could not copy build output to output directory: %w", err)
	}

	if err = os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("could not create output directory: %w", err)
	}

	if err = copyFS(os.DirFS(filepath.Join(buildDir, "out")), outputDir); err != nil {
		return fmt.Errorf("could not copy build output to output directory: %w", err)
	}

	// Copy doc_classes
	docsSrc := filepath.Join(buildDir, "doc_classes")
	if _, err := os.Stat(docsSrc); err == nil {
		docsDest := filepath.Join(outputDir, "doc_classes")
		if err := os.MkdirAll(docsDest, 0755); err != nil {
			return fmt.Errorf("could not create docs directory: %w", err)
		}
		if err := copyFS(os.DirFS(docsSrc), docsDest); err != nil {
			return fmt.Errorf("could not copy doc files: %w", err)
		}
	}

	return nil
}

func copyFS(src fs.FS, dst string) error {
	return fs.WalkDir(src, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, path)
		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		srcFile, err := src.Open(path)
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
