package jbtf

import (
	"io"
)

// Catch-all helper for serializing arbitrary binary data
type Bytes struct {
	Data []byte
}

func (pi *Bytes) Deserialize(r io.Reader) (err error) {
	pi.Data, err = io.ReadAll(r)
	return err
}

func (pi Bytes) Serialize(w io.Writer) error {
	_, err := w.Write(pi.Data)
	return err
}
