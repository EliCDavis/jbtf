package jbtf

import "io"

type Serializable interface {
	Deserialize(io.Reader) error
	Serialize(io.Writer) error
}
