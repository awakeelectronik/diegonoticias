//go:build vips

package images

import (
	"sync"

	vips "github.com/davidbyttow/govips/v2/vips"
)

var vipsOnce sync.Once

// initRuntime no inicializa vips al arrancar.
// La inicialización ocurre de forma lazy la primera vez que se procesa una imagen.
func initRuntime() (ShutdownFunc, error) {
	return func() {
		vips.Shutdown()
	}, nil
}

// EnsureVips inicializa libvips la primera vez que se llama.
// Es seguro llamarla múltiples veces desde distintas goroutines.
func EnsureVips() {
	vipsOnce.Do(func() {
		vips.Startup(&vips.Config{
			MaxCacheSize:  0,
			MaxCacheMem:   0,
			MaxCacheFiles: 0,
		})
	})
}
