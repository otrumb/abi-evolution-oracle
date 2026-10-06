package experiment

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type generatedField struct{ Type, Field string }

func compileGeneratedPair(root, tool, oldPath, newPath string) (bool, error) {
	temp, err := os.MkdirTemp("", "abi-generated-")
	if err != nil {
		return false, fmt.Errorf("create generated temp: %w", err)
	}
	defer os.RemoveAll(temp)
	oldFile, newFile := filepath.Join(temp, "old.go"), filepath.Join(temp, "new.go")
	if err := generateBinding(root, tool, oldPath, oldFile); err != nil {
		return false, err
	}
	if err := generateBinding(root, tool, newPath, newFile); err != nil {
		return false, err
	}
	oldFields, err := exportedFields(oldFile)
	if err != nil {
		return false, err
	}
	newFields, err := exportedFields(newFile)
	if err != nil {
		return false, err
	}
	candidate := firstRemovedField(oldFields, newFields)
	consumer := "package binding\n"
	if candidate.Type != "" {
		consumer += fmt.Sprintf("func consume(value %s) { _ = value.%s }\n", candidate.Type, candidate.Field)
	}
	consumerFile := filepath.Join(temp, "consumer.go")
	if err := os.WriteFile(consumerFile, []byte(consumer), 0644); err != nil {
		return false, fmt.Errorf("write consumer: %w", err)
	}
	if err := compileGenerated(root, oldFile, consumerFile); err != nil {
		return false, fmt.Errorf("compile old consumer: %w", err)
	}
	newErr := compileGenerated(root, newFile, consumerFile)
	return newErr != nil, nil
}

func generateBinding(root, tool, path, output string) error {
	command := exec.Command(tool, "--abi", filepath.Join(root, filepath.FromSlash(path)), "--pkg", "binding", "--type", "Fixture", "--out", output)
	if data, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("abigen: %s: %w", string(data), err)
	}
	return nil
}

func compileGenerated(root string, files ...string) error {
	args := append([]string{"test", "-mod=mod"}, files...)
	command := exec.Command("go", args...)
	command.Dir = root
	command.Env = append(os.Environ(), "CGO_ENABLED=0", "GOTOOLCHAIN=local")
	if data, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("go test: %s: %w", string(data), err)
	}
	return nil
}

func exportedFields(path string) (map[string]bool, error) {
	parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return nil, fmt.Errorf("parse generated: %w", err)
	}
	result := map[string]bool{}
	ast.Inspect(parsed, func(node ast.Node) bool {
		declaration, ok := node.(*ast.TypeSpec)
		if !ok {
			return true
		}
		structure, ok := declaration.Type.(*ast.StructType)
		if !ok {
			return true
		}
		for _, field := range structure.Fields.List {
			for _, name := range field.Names {
				if name.IsExported() {
					result[declaration.Name.Name+"."+name.Name] = true
				}
			}
		}
		return true
	})
	return result, nil
}

func firstRemovedField(oldFields, newFields map[string]bool) generatedField {
	var removed []string
	for field := range oldFields {
		if !newFields[field] {
			removed = append(removed, field)
		}
	}
	sort.Strings(removed)
	if len(removed) == 0 {
		return generatedField{}
	}
	parts := strings.SplitN(removed[0], ".", 2)
	return generatedField{parts[0], parts[1]}
}
