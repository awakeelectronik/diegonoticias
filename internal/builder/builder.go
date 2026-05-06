package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

type BuildResult struct {
	Status    string
	Error     string
	StartedAt time.Time
	EndedAt   time.Time
	Duration  time.Duration
}

type Builder struct {
	siteDir  string
	hugoBin  string
	mu       sync.Mutex
	lastBuild BuildResult
}

func New(siteDir, hugoBin string) *Builder {
	return &Builder{
		siteDir: siteDir,
		hugoBin: hugoBin,
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
	cmd := exec.Command(b.hugoBin, "--minify")
	cmd.Dir = b.siteDir
	cmd.Env = os.Environ()
	if out, err := cmd.CombinedOutput(); err != nil {
		res.Status = "error"
		res.Error = fmt.Sprintf("%v: %s", err, string(out))
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

