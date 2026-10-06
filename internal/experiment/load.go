package experiment

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"local/abi-evolution-oracle-validation/internal/corpus"
)

func loadEntries(root string) ([]corpus.Entry, error) {
	var result []corpus.Entry
	for _, class := range []string{"directional", "names", "events", "collisions", "structural"} {
		file, err := os.Open(filepath.Join(root, "corpus", "manifests", class+".jsonl"))
		if err != nil {
			return nil, fmt.Errorf("open manifest: %w", err)
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			var value corpus.Entry
			if err := json.Unmarshal(scanner.Bytes(), &value); err != nil {
				file.Close()
				return nil, fmt.Errorf("decode manifest: %w", err)
			}
			result = append(result, value)
		}
		if err := scanner.Err(); err != nil {
			file.Close()
			return nil, fmt.Errorf("scan manifest: %w", err)
		}
		if err := file.Close(); err != nil {
			return nil, fmt.Errorf("close manifest: %w", err)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}

func readABI(root, path string) ([]byte, error) {
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
	if err != nil {
		return nil, fmt.Errorf("read ABI: %w", err)
	}
	return data, nil
}
