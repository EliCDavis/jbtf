package jbtf

import (
	"errors"
	"image"
	"image/png"
	"io"
)

// Helper struct for serializing images
type Png struct {
	Image image.Image
}

func (pi *Png) Deserialize(r io.Reader) (err error) {
	pi.Image, _, err = image.Decode(r)
	return err
}

func (pi Png) Serialize(w io.Writer) error {
	if pi.Image == nil {
		return errors.New("can not serialize nil image.")
	}
	return png.Encode(w, pi.Image)
}
