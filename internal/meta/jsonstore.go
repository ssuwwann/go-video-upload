package meta

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

type JSONStore struct {
	root string
}

func NewJSONStore(root string) *JSONStore {
	return &JSONStore{root: root}
}

func (s *JSONStore) pathFor(id string) string {
	return filepath.Join(s.root, fmt.Sprintf("%s.json", id))
}

func (s *JSONStore) Create(m Metadata) error {
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	p := s.pathFor(m.ID)
	if err := writeFileAtomic(p, m); err != nil {
		return err
	}
	return nil
}

func (s *JSONStore) Get(id string) (Metadata, error) {
	p := s.pathFor(id)
	b, err := os.ReadFile(p)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Metadata{}, fs.ErrNotExist
		}
		return Metadata{}, err
	}
	var m Metadata
	if err := json.Unmarshal(b, &m); err != nil {
		return Metadata{}, err
	}
	return m, nil
}

func (s *JSONStore) Update(m Metadata) error {
	m.UpdatedAt = time.Now()
	p := s.pathFor(m.ID)
	return writeFileAtomic(p, m)
}

func (s *JSONStore) List() ([]Metadata, error) {
	entries, err := os.ReadDir(s.root)
	if err != nil {
		return nil, err
	}
	out := make([]Metadata, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		b, err := os.ReadFile(filepath.Join(s.root, e.Name()))
		if err != nil {
			continue
		}
		var m Metadata
		if err := json.Unmarshal(b, &m); err == nil {
			out = append(out, m)
		}
	}
	return out, nil
}

// writeFileAtomic writes JSON to a temp file then renames it into place.
func writeFileAtomic(dest string, v any) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	tmp := dest + ".tmp"
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	// On Windows, rename over existing fails; remove first.
	_ = os.Remove(dest)
	return os.Rename(tmp, dest)
}
