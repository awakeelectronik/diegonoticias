//go:build vips

package images

import (
	"fmt"
	"os"
	"path/filepath"

	vips "github.com/davidbyttow/govips/v2/vips"
)

func processWithVips(input []byte, diskBase string, cfg Config) error {
	img, err := vips.NewImageFromBuffer(input)
	if err != nil {
		return fmt.Errorf("decode image: %w", err)
	}
	defer img.Close()
	img.RemoveMetadata()
	origW := img.Width()
	for _, w := range cfg.Sizes {
		if w > origW {
			continue
		}
		scale := float64(w) / float64(origW)
		thumb, err := img.Copy()
		if err != nil {
			return err
		}
		if err := thumb.Resize(scale, vips.KernelLanczos3); err != nil {
			thumb.Close()
			return err
		}
		avifBytes, _, err := thumb.ExportAvif(&vips.AvifExportParams{Quality: cfg.AvifQuality})
		if err != nil {
			thumb.Close()
			return err
		}
		if err := atomicWriteFile(diskBase+fmt.Sprintf("-%d.avif", w), avifBytes); err != nil {
			thumb.Close()
			return err
		}
		webpBytes, _, err := thumb.ExportWebp(&vips.WebpExportParams{Quality: cfg.WebpQuality})
		if err != nil {
			thumb.Close()
			return err
		}
		if err := atomicWriteFile(diskBase+fmt.Sprintf("-%d.webp", w), webpBytes); err != nil {
			thumb.Close()
			return err
		}
		thumb.Close()
	}
	return nil
}

func atomicWriteFile(path string, content []byte) error {
	tmp := path + ".tmp"
	if err := atomicMkdir(path); err != nil {
		return err
	}
	if err := osWriteFile(tmp, content); err != nil {
		return err
	}
	return osRename(tmp, path)
}

var (
	atomicMkdir = func(path string) error { return os.MkdirAll(filepath.Dir(path), 0o755) }
	osWriteFile = func(path string, content []byte) error { return os.WriteFile(path, content, 0o644) }
	osRename    = os.Rename
)

