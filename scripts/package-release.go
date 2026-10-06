//go:build ignore

package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const releaseVersion = "v0.1.0"

var archiveTime = time.Date(1980, 1, 1, 0, 0, 0, 0, time.UTC)

type target struct {
	goos       string
	binaryName string
	archive    string
}

func main() {
	if len(os.Args) != 3 || os.Args[1] != releaseVersion {
		fmt.Fprintf(os.Stderr, "usage: go run ./scripts/package-release.go %s DIST_DIR\n", releaseVersion)
		os.Exit(2)
	}
	dist, err := filepath.Abs(os.Args[2])
	if err != nil {
		fail(err)
	}
	if err := os.RemoveAll(dist); err != nil {
		fail(err)
	}
	if err := os.MkdirAll(dist, 0o755); err != nil {
		fail(err)
	}

	targets := []target{
		{goos: "linux", binaryName: "abi-evolution-oracle", archive: "abi-evolution-oracle_0.1.0_linux_amd64.tar.gz"},
		{goos: "windows", binaryName: "abi-evolution-oracle.exe", archive: "abi-evolution-oracle_0.1.0_windows_amd64.zip"},
	}
	for _, current := range targets {
		if err := packageTarget(dist, current); err != nil {
			fail(err)
		}
	}
	if err := writeChecksums(dist, targets); err != nil {
		fail(err)
	}
}

func packageTarget(dist string, current target) error {
	staging, err := os.MkdirTemp("", "abi-evolution-oracle-package-")
	if err != nil {
		return fmt.Errorf("create staging directory: %w", err)
	}
	defer os.RemoveAll(staging)

	binary := filepath.Join(staging, current.binaryName)
	command := exec.Command("go", "build", "-buildvcs=false", "-trimpath", "-ldflags=-buildid=", "-o", binary, "./cmd/abi-evolution-oracle")
	command.Env = append(os.Environ(), "CGO_ENABLED=0", "GOARCH=amd64", "GOOS="+current.goos, "GOTOOLCHAIN=local")
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("build %s/amd64: %w", current.goos, err)
	}
	files := []string{binary, "LICENSE", "README.md"}
	destination := filepath.Join(dist, current.archive)
	if current.goos == "linux" {
		return writeTarGzip(destination, files)
	}
	return writeZip(destination, files)
}

func writeTarGzip(destination string, files []string) error {
	output, err := os.Create(destination)
	if err != nil {
		return err
	}
	compressor := gzip.NewWriter(output)
	compressor.Header.ModTime = archiveTime
	compressor.Header.OS = 255
	archive := tar.NewWriter(compressor)
	for _, path := range files {
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		mode := int64(0o644)
		if filepath.Base(path) == "abi-evolution-oracle" {
			mode = 0o755
		}
		header := &tar.Header{Name: filepath.Base(path), Mode: mode, Size: int64(len(data)), ModTime: archiveTime, AccessTime: archiveTime, ChangeTime: archiveTime, Format: tar.FormatPAX}
		if err := archive.WriteHeader(header); err != nil {
			return err
		}
		if _, err := archive.Write(data); err != nil {
			return err
		}
	}
	if err := archive.Close(); err != nil {
		return err
	}
	if err := compressor.Close(); err != nil {
		return err
	}
	return output.Close()
}

func writeZip(destination string, files []string) error {
	output, err := os.Create(destination)
	if err != nil {
		return err
	}
	archive := zip.NewWriter(output)
	for _, path := range files {
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		header := &zip.FileHeader{Name: filepath.Base(path), Method: zip.Deflate}
		header.Modified = archiveTime
		header.SetMode(0o644)
		if strings.HasSuffix(path, ".exe") {
			header.SetMode(0o755)
		}
		writer, createErr := archive.CreateHeader(header)
		if createErr != nil {
			return createErr
		}
		if _, err := writer.Write(data); err != nil {
			return err
		}
	}
	if err := archive.Close(); err != nil {
		return err
	}
	return output.Close()
}

func writeChecksums(dist string, targets []target) error {
	output, err := os.Create(filepath.Join(dist, "SHA256SUMS"))
	if err != nil {
		return err
	}
	defer output.Close()
	for _, current := range targets {
		path := filepath.Join(dist, current.archive)
		input, openErr := os.Open(path)
		if openErr != nil {
			return openErr
		}
		hash := sha256.New()
		_, copyErr := io.Copy(hash, input)
		closeErr := input.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		if _, err := fmt.Fprintf(output, "%x  %s\n", hash.Sum(nil), current.archive); err != nil {
			return err
		}
	}
	return nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
