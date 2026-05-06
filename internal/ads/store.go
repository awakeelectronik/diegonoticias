package ads

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

type Banner struct {
	ID        string    `json:"id" toml:"id"`
	Title     string    `json:"title" toml:"title"`
	ImagePath string    `json:"imagePath" toml:"imagePath"`
	Active    bool      `json:"active" toml:"active"`
	Slot      int       `json:"slot" toml:"slot"`
	CreatedAt time.Time `json:"createdAt" toml:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" toml:"updatedAt"`
}

type dataFile struct {
	Banners []Banner `json:"banners"`
}

type Store struct {
	path string
}

func New(path string) *Store { return &Store{path: path} }

func (s *Store) List() ([]Banner, error) {
	data, err := s.read()
	if err != nil {
		return nil, err
	}
	return data.Banners, nil
}

func (s *Store) Create(b Banner) (Banner, error) {
	all, err := s.read()
	if err != nil {
		return Banner{}, err
	}
	b.ID = uuid.NewString()
	now := time.Now()
	b.CreatedAt = now
	b.UpdatedAt = now
	if err := Validate(b, all.Banners, ""); err != nil {
		return Banner{}, err
	}
	all.Banners = append(all.Banners, b)
	return b, s.write(all)
}

func (s *Store) Update(id string, in Banner) (Banner, error) {
	all, err := s.read()
	if err != nil {
		return Banner{}, err
	}
	for i, b := range all.Banners {
		if b.ID != id {
			continue
		}
		in.ID = id
		in.CreatedAt = b.CreatedAt
		in.UpdatedAt = time.Now()
		if err := Validate(in, all.Banners, id); err != nil {
			return Banner{}, err
		}
		all.Banners[i] = in
		return in, s.write(all)
	}
	return Banner{}, os.ErrNotExist
}

func (s *Store) Delete(id string) error {
	all, err := s.read()
	if err != nil {
		return err
	}
	out := make([]Banner, 0, len(all.Banners))
	found := false
	for _, b := range all.Banners {
		if b.ID == id {
			found = true
			continue
		}
		out = append(out, b)
	}
	if !found {
		return os.ErrNotExist
	}
	all.Banners = out
	return s.write(all)
}

func (s *Store) read() (dataFile, error) {
	b, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return dataFile{Banners: []Banner{}}, nil
		}
		return dataFile{}, err
	}
	var d dataFile
	if err := json.Unmarshal(b, &d); err != nil {
		return dataFile{}, err
	}
	return d, nil
}

func (s *Store) write(v dataFile) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

var ErrInvalid = errors.New("banner inválido")

