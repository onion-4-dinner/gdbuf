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

func (gde *GDExtensionBuilder) ExtractNanopbGenerator(dst string) error {
	genFS, err := fs.Sub(buildEnvFS, "buildenv/nanopb/generator")
	if err != nil {
		return err
	}
	return copyFS(genFS, dst)
}

func (gde *GDExtensionBuilder) Build(generatedCppSourceDir, outputDir string, generateOnly bool) error {
	// Determine build directory: UserCacheDir/gdbuf
	userCacheDir, err := os.UserCacheDir()
	var buildDir string
	if err != nil {
		// Fallback
		buildDir = filepath.Join(".", ".gdbuf_cache")
	} else {
		buildDir = filepath.Join(userCacheDir, "gdbuf")
	}

	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return fmt.Errorf("could not make build directory: %w", err)
	}

	gde.logger.Info("preparing gdextension build environment", "build_dir", buildDir)

	buildEnv, err := fs.Sub(buildEnvFS, "buildenv")
	if err != nil {
		return fmt.Errorf("could not get build environment fs: %w", err)
	}

	if err = copyFS(buildEnv, buildDir); err != nil {
		return fmt.Errorf("could not copy build environment to build directory: %w", err)
	}

	if err = copyFS(os.DirFS(generatedCppSourceDir), buildDir); err != nil {
		return fmt.Errorf("could not copy custom build files to build directory: %w", err)
	}

	// Copy doc_classes from source to build dir if they exist (to be packaged later)
	docsSrc := filepath.Join(generatedCppSourceDir, "doc_classes")
	if _, err := os.Stat(docsSrc); err == nil {
		docsDest := filepath.Join(buildDir, "doc_classes")
		if err := os.MkdirAll(docsDest, 0755); err != nil {
			return fmt.Errorf("could not create docs directory in build dir: %w", err)
		}
		if err := copyFS(os.DirFS(docsSrc), docsDest); err != nil {
			return fmt.Errorf("could not copy doc files to build dir: %w", err)
		}
	}

	if generateOnly {
		gde.logger.Info("skipping build step as --generate-only was provided")
		return nil
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

	// Also copy doc_classes to the final output (addon)
	if _, err := os.Stat(docsSrc); err == nil {
		docsDest := filepath.Join(outputDir, "doc_classes")
		if err := os.MkdirAll(docsDest, 0755); err != nil {
			return fmt.Errorf("could not create docs directory in output: %w", err)
		}
		if err := copyFS(os.DirFS(docsSrc), docsDest); err != nil {
			return fmt.Errorf("could not copy doc files to output: %w", err)
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
