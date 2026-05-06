//go:build !vips

package images

import "errors"

func processWithVips(_ []byte, _ string, _ Config) error {
	return errors.New("pipeline de imágenes no disponible: compila con -tags vips")
}

