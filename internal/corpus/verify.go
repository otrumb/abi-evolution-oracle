package corpus

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Hashes struct {
	Old   string `json:"old"`
	New   string `json:"new"`
	Probe string `json:"probe"`
}
type Entry struct {
	ID       string          `json:"id"`
	Class    string          `json:"class"`
	Old      string          `json:"old"`
	New      string          `json:"new"`
	Probe    string          `json:"probe"`
	License  string          `json:"license"`
	Origin   string          `json:"origin"`
	Expected map[string]bool `json:"expected_observations"`
	Hashes   Hashes          `json:"sha256"`
}
type Summary struct {
	Total       int `json:"total"`
	Directional int `json:"directional"`
	Names       int `json:"names"`
	Events      int `json:"events"`
	Collisions  int `json:"collisions"`
	Structural  int `json:"structural"`
}

func Verify(root string) (Summary, error) {
	var entries []Entry
	for _, class := range []string{"directional", "names", "events", "collisions", "structural"} {
		file, err := os.Open(filepath.Join(root, "corpus", "manifests", class+".jsonl"))
		if err != nil {
			return Summary{}, fmt.Errorf("open %s manifest: %w", class, err)
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			var entry Entry
			if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
				file.Close()
				return Summary{}, fmt.Errorf("parse %s manifest: %w", class, err)
			}
			entries = append(entries, entry)
		}
		if err := scanner.Err(); err != nil {
			file.Close()
			return Summary{}, fmt.Errorf("scan %s manifest: %w", class, err)
		}
		if err := file.Close(); err != nil {
			return Summary{}, fmt.Errorf("close %s manifest: %w", class, err)
		}
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].ID < entries[j].ID })
	seen := make(map[string]bool, len(entries))
	summary := Summary{Total: len(entries)}
	for _, entry := range entries {
		if seen[entry.ID] {
			return Summary{}, fmt.Errorf("duplicate fixture %s", entry.ID)
		}
		seen[entry.ID] = true
		if entry.License != "CC0-1.0" || entry.Origin != "original" || len(entry.Expected) == 0 {
			return Summary{}, fmt.Errorf("invalid provenance or expectation for %s", entry.ID)
		}
		for path, want := range map[string]string{entry.Old: entry.Hashes.Old, entry.New: entry.Hashes.New, entry.Probe: entry.Hashes.Probe} {
			if filepath.IsAbs(path) || strings.Contains(path, "..") {
				return Summary{}, fmt.Errorf("unsafe path for %s", entry.ID)
			}
			data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
			if err != nil {
				return Summary{}, fmt.Errorf("read %s: %w", path, err)
			}
			sum := sha256.Sum256(data)
			if hex.EncodeToString(sum[:]) != want {
				return Summary{}, fmt.Errorf("hash mismatch for %s", path)
			}
		}
		switch entry.Class {
		case "directional":
			summary.Directional++
		case "names":
			summary.Names++
		case "events":
			summary.Events++
		case "collisions":
			summary.Collisions++
		case "structural":
			summary.Structural++
		default:
			return Summary{}, fmt.Errorf("unknown class %s", entry.Class)
		}
	}
	if summary != (Summary{50, 12, 12, 12, 8, 6}) {
		return Summary{}, fmt.Errorf("wrong corpus split: %+v", summary)
	}
	return summary, nil
}
