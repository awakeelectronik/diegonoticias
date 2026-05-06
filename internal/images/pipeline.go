package images

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Config struct {
	UploadsRoot    string
	MaxUploadBytes int64
	Sizes          []int
	AvifQuality    int
	WebpQuality    int
}

type Pipeline struct {
	cfg Config
}

func NewPipeline(cfg Config) *Pipeline {
	if cfg.MaxUploadBytes == 0 {
		cfg.MaxUploadBytes = 8 * 1024 * 1024
	}
	if len(cfg.Sizes) == 0 {
		cfg.Sizes = []int{320, 640, 1024, 1600}
	}
	if cfg.AvifQuality == 0 {
		cfg.AvifQuality = 55
	}
	if cfg.WebpQuality == 0 {
		cfg.WebpQuality = 72
	}
	return &Pipeline{cfg: cfg}
}

func (p *Pipeline) ProcessArticleImage(input []byte) (string, error) {
	if int64(len(input)) > p.cfg.MaxUploadBytes {
		return "", fmt.Errorf("imagen excede máximo de %d bytes", p.cfg.MaxUploadBytes)
	}
	now := time.Now()
	hash, err := randomHex(8)
	if err != nil {
		return "", err
	}
	base := fmt.Sprintf("images/%04d/%02d/%s", now.Year(), now.Month(), hash)
	diskBase := filepath.Join(p.cfg.UploadsRoot, base)
	if err := os.MkdirAll(filepath.Dir(diskBase), 0o755); err != nil {
		return "", err
	}
	if err := processWithVips(input, diskBase, p.cfg); err != nil {
		return "", err
	}
	smallest := p.cfg.Sizes[0]
	check := fmt.Sprintf("%s-%d.webp", diskBase, smallest)
	if _, err := os.Stat(check); err != nil {
		return "", fmt.Errorf("pipeline no generó variantes (%s): %w", check, err)
	}
	return "/" + base, nil
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

