package jbtf_test

import (
	"bytes"
	"testing"

	"github.com/EliCDavis/jbtf"
	"github.com/stretchr/testify/assert"
)

func TestBytesSerializer(t *testing.T) {
	bytesSerializer := jbtf.Bytes{[]byte{1, 2, 3, 4}}
	bytesSerializerBack := jbtf.Bytes{}

	buf := &bytes.Buffer{}
	assert.NoError(t, bytesSerializer.Serialize(buf))
	assert.NoError(t, bytesSerializerBack.Deserialize(buf))

	assert.Equal(t, bytesSerializer.Data, bytesSerializerBack.Data)
}
