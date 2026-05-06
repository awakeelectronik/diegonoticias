//go:build !vips

package images

func initRuntime() (ShutdownFunc, error) {
	return func() {}, nil
}

