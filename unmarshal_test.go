package jbtf_test

import (
	"testing"

	"github.com/EliCDavis/jbtf"
	"github.com/stretchr/testify/assert"
)

func TestParseJsonWithBuffers(t *testing.T) {

	// ARRANGE ================================================================
	type TestStruct struct {
		A string
		B *jbtfSerializableStruct
	}

	json := []byte(`{
		"A": "Something",
		"$B": 0  
	}`)

	buffers := []jbtf.Buffer{
		{
			ByteLength: 4,
			URI:        "data:application/octet-stream;base64,OTAAAA==",
		},
	}

	bufferViews := []jbtf.BufferView{
		{
			Buffer:     0,
			ByteOffset: 0,
			ByteLength: 4,
		},
	}

	// ACT ====================================================================
	result, err := jbtf.ParseJsonUsingBuffers[TestStruct](buffers, bufferViews, json)

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.Equal(t, "Something", result.A)
	assert.Equal(t, int32(12345), result.B.data)
}
