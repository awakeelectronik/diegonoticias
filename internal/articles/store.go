package articles

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
	"gopkg.in/yaml.v2"
)

type Store struct {
	contentDir string
}

func NewStore(contentDir string) *Store {
	return &Store{contentDir: contentDir}
}

func (s *Store) List() ([]Article, error) {
	entries, err := os.ReadDir(s.contentDir)
	if err != nil {
		return nil, fmt.Errorf("read content dir: %w", err)
	}
	out := make([]Article, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		a, err := s.Get(strings.TrimSuffix(e.Name(), ".md"))
		if err != nil {
			continue
		}
		out = append(out, *a)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Date.After(out[j].Date) })
	return out, nil
}

func (s *Store) Exists(slug string) bool {
	_, err := os.Stat(filepath.Join(s.contentDir, slug+".md"))
	return err == nil
}

func (s *Store) Get(slug string) (*Article, error) {
	path := filepath.Join(s.contentDir, slug+".md")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read article: %w", err)
	}
	var fm Frontmatter
	body, err := frontmatter.Parse(bytes.NewReader(b), &fm)
	if err != nil {
		return nil, fmt.Errorf("parse frontmatter: %w", err)
	}
	return &Article{
		Frontmatter: fm,
		Body:        strings.TrimSpace(string(body)),
	}, nil
}

func (s *Store) Create(a *Article) error {
	if strings.TrimSpace(a.Title) == "" {
		return errors.New("título requerido")
	}
	if strings.TrimSpace(a.Slug) == "" {
		a.Slug = GenerateSlug(a.Title, s.Exists)
	}
	if s.Exists(a.Slug) {
		return errors.New("slug ya existe")
	}
	if a.Date.IsZero() {
		a.Date = time.Now()
	}
	return s.writeFile(a.Slug, a)
}

func (s *Store) Update(slug string, a *Article) error {
	if !s.Exists(slug) {
		return os.ErrNotExist
	}
	if strings.TrimSpace(a.Slug) == "" {
		a.Slug = slug
	}
	if a.Date.IsZero() {
		a.Date = time.Now()
	}
	if err := s.writeFile(a.Slug, a); err != nil {
		return err
	}
	if a.Slug != slug {
		_ = os.Remove(filepath.Join(s.contentDir, slug+".md"))
	}
	return nil
}

func (s *Store) Delete(slug string) error {
	return os.Remove(filepath.Join(s.contentDir, slug+".md"))
}

func (s *Store) writeFile(slug string, a *Article) error {
	if err := os.MkdirAll(s.contentDir, 0o755); err != nil {
		return err
	}
	fm, err := yaml.Marshal(a.Frontmatter)
	if err != nil {
		return fmt.Errorf("marshal frontmatter: %w", err)
	}
	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(fm)
	buf.WriteString("---\n\n")
	buf.WriteString(strings.TrimSpace(a.Body))
	buf.WriteString("\n")
	dst := filepath.Join(s.contentDir, slug+".md")
	tmp := dst + ".tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, dst)
}

