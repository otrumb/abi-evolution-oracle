//go:build ignore

package main

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

var archives = []string{
	"abi-evolution-oracle_0.1.0_linux_amd64.tar.gz",
	"abi-evolution-oracle_0.1.0_windows_amd64.zip",
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: go run ./scripts/verify-release.go DIST_DIR")
		os.Exit(2)
	}
	dist, err := filepath.Abs(os.Args[1])
	if err != nil {
		fail(err)
	}
	if err := verifyChecksums(dist); err != nil {
		fail(err)
	}
	for _, archive := range archives {
		if err := verifyArchive(dist, archive); err != nil {
			fail(err)
		}
	}
}

func verifyChecksums(dist string) error {
	file, err := os.Open(filepath.Join(dist, "SHA256SUMS"))
	if err != nil {
		return err
	}
	defer file.Close()
	found := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 || !contains(archives, fields[1]) || found[fields[1]] {
			return fmt.Errorf("invalid checksum entry %q", scanner.Text())
		}
		expected, err := hex.DecodeString(fields[0])
		if err != nil || len(expected) != sha256.Size {
			return fmt.Errorf("invalid checksum for %s", fields[1])
		}
		data, err := os.ReadFile(filepath.Join(dist, fields[1]))
		if err != nil {
			return err
		}
		actual := sha256.Sum256(data)
		if !equal(expected, actual[:]) {
			return fmt.Errorf("checksum mismatch for %s", fields[1])
		}
		found[fields[1]] = true
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if len(found) != len(archives) {
		return fmt.Errorf("checksum manifest has %d archives", len(found))
	}
	return nil
}

func verifyArchive(dist, name string) error {
	destination, err := os.MkdirTemp("", "abi-evolution-oracle-verify-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(destination)
	if strings.HasSuffix(name, ".zip") {
		err = extractZip(filepath.Join(dist, name), destination)
	} else {
		err = extractTarGzip(filepath.Join(dist, name), destination)
	}
	if err != nil {
		return err
	}
	binary := "abi-evolution-oracle"
	if strings.Contains(name, "windows") {
		binary += ".exe"
	}
	expected := []string{"LICENSE", "README.md", binary}
	entries, err := os.ReadDir(destination)
	if err != nil {
		return err
	}
	if len(entries) != len(expected) {
		return fmt.Errorf("%s contains %d entries", name, len(entries))
	}
	for _, file := range expected {
		if _, err := os.Stat(filepath.Join(destination, file)); err != nil {
			return fmt.Errorf("%s missing %s", name, file)
		}
	}
	return smokeBinary(filepath.Join(destination, binary), name)
}

func extractZip(path, destination string) error {
	archive, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	defer archive.Close()
	for _, file := range archive.File {
		if filepath.Base(file.Name) != file.Name {
			return fmt.Errorf("unsafe zip entry %q", file.Name)
		}
		input, err := file.Open()
		if err != nil {
			return err
		}
		if err := writeFile(filepath.Join(destination, file.Name), input, file.Mode()); err != nil {
			input.Close()
			return err
		}
		input.Close()
	}
	return nil
}

func extractTarGzip(path, destination string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	compressed, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer compressed.Close()
	archive := tar.NewReader(compressed)
	for {
		header, err := archive.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if filepath.Base(header.Name) != header.Name || !header.FileInfo().Mode().IsRegular() {
			return fmt.Errorf("unsafe tar entry %q", header.Name)
		}
		if err := writeFile(filepath.Join(destination, header.Name), archive, header.FileInfo().Mode()); err != nil {
			return err
		}
	}
}

func writeFile(path string, reader io.Reader, mode os.FileMode) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(file, reader)
	closeErr := file.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func smokeBinary(path, archive string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	isWindows := strings.Contains(archive, "windows")
	validFormat := len(data) >= 4 && ((!isWindows && string(data[:4]) == "\x7fELF") || (isWindows && string(data[:2]) == "MZ"))
	if !validFormat {
		return fmt.Errorf("invalid binary format in %s", archive)
	}
	if (runtime.GOOS == "windows") != isWindows {
		return nil
	}
	command := exec.Command(path)
	output, err := command.CombinedOutput()
	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.ExitCode() != 2 || !strings.Contains(string(output), "usage: abi-evolution-oracle") {
		return fmt.Errorf("native smoke failed for %s: %s", archive, output)
	}
	return nil
}

func contains(values []string, candidate string) bool {
	for _, value := range values {
		if value == candidate {
			return true
		}
	}
	return false
}

func equal(left, right []byte) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
