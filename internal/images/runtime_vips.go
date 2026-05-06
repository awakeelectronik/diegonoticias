//go:build vips

package images

import vips "github.com/davidbyttow/govips/v2/vips"

func initRuntime() (ShutdownFunc, error) {
	vips.Startup(&vips.Config{
		MaxCacheSize: 0,
		MaxCacheMem:  0,
	})
	return func() {
		vips.Shutdown()
	}, nil
}

