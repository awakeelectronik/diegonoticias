package builder

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/awakeelectronik/diegonoticias/internal/ads"
	"github.com/awakeelectronik/diegonoticias/internal/settings"
)

type BuildResult struct {
	Status    string
	Error     string
	StartedAt time.Time
	EndedAt   time.Time
	Duration  time.Duration
}

type Builder struct {
	siteDir   string
	dataDir   string
	hugoBin   string
	pagefindBin string
	mu        sync.Mutex
	lastBuild BuildResult
}

func New(siteDir, dataDir, hugoBin, pagefindBin string) *Builder {
	return &Builder{
		siteDir: siteDir,
		dataDir: dataDir,
		hugoBin: hugoBin,
		pagefindBin: pagefindBin,
	}
}

func (b *Builder) Build() BuildResult {
	b.mu.Lock()
	defer b.mu.Unlock()
	start := time.Now()
	res := BuildResult{
		Status:    "ok",
		StartedAt: start,
	}
	publicDir := filepath.Join(b.siteDir, "public")
	_ = os.MkdirAll(publicDir, 0o755)
	if err := b.exportData(); err != nil {
		res.Status = "error"
		res.Error = err.Error()
		res.EndedAt = time.Now()
		res.Duration = res.EndedAt.Sub(start)
		b.lastBuild = res
		return res
	}
	cmd := exec.Command(b.hugoBin, "--minify")
	cmd.Dir = b.siteDir
	cmd.Env = os.Environ()
	if out, err := cmd.CombinedOutput(); err != nil {
		res.Status = "error"
		res.Error = fmt.Sprintf("%v: %s", err, string(out))
		res.EndedAt = time.Now()
		res.Duration = res.EndedAt.Sub(start)
		b.lastBuild = res
		return res
	}
	pagefindCmd := exec.Command(b.pagefindBin, "--site", "public")
	pagefindCmd.Dir = b.siteDir
	pagefindCmd.Env = os.Environ()
	if out, err := pagefindCmd.CombinedOutput(); err != nil {
		res.Status = "error"
		res.Error = fmt.Sprintf("pagefind: %v: %s", err, string(out))
	}
	res.EndedAt = time.Now()
	res.Duration = res.EndedAt.Sub(start)
	b.lastBuild = res
	return res
}

func (b *Builder) LastBuild() BuildResult {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.lastBuild
}

func (b *Builder) exportData() error {
	siteDataDir := filepath.Join(b.siteDir, "data")
	if err := os.MkdirAll(siteDataDir, 0o755); err != nil {
		return err
	}
	settingsPath := filepath.Join(b.dataDir, "settings.json")
	adsPath := filepath.Join(b.dataDir, "ads.json")

	st := settings.Settings{
		SiteName:        "Diego Noticias",
		SiteDescription: "Noticias breves y al grano.",
		SiteURL:         "https://diegonoticias.com",
		DefaultOgImage:  "/og-default.jpg",
	}
	if bts, err := os.ReadFile(settingsPath); err == nil {
		_ = json.Unmarshal(bts, &st)
	}
	if err := writeToml(filepath.Join(siteDataDir, "site.toml"), st); err != nil {
		return err
	}

	var adFile struct {
		Banners []ads.Banner `json:"banners"`
	}
	if bts, err := os.ReadFile(adsPath); err == nil {
		_ = json.Unmarshal(bts, &adFile)
	}
	active := make([]ads.Banner, 0, len(adFile.Banners))
	for _, b := range adFile.Banners {
		if b.Active {
			active = append(active, b)
		}
	}
	return writeToml(filepath.Join(siteDataDir, "ads.toml"), struct {
		Banners []ads.Banner `toml:"banners"`
	}{Banners: active})
}

func writeToml(path string, v any) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(v)
}

