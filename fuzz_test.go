package jbtf_test

import (
	"testing"

	"github.com/EliCDavis/jbtf"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshal_MalformedBuffers(t *testing.T) {
	tests := map[string]string{
		"buffer without uri":     `{"Buffers":[{}]}`,
		"buffer with bad base64": `{"buffers":[{"uri":"data:application/octet-stream;base64,!!!"}]}`,
	}

	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := jbtf.Unmarshal[embeddedBinaryStruct]([]byte(input))
			assert.Error(t, err)

			_, err = jbtf.NewDecoder([]byte(input))
			assert.Error(t, err)
		})
	}
}

func TestUnmarshal_MalformedBufferReferences(t *testing.T) {
	const buffers = `"buffers":[{"byteLength":4,"uri":"data:application/octet-stream;base64,OTAAAA=="}]`
	tests := map[string]string{
		"non-numeric index":       `{` + buffers + `,"bufferViews":[{"buffer":0,"byteLength":4}],"data":{"$Serializable":"0"}}`,
		"negative index":          `{` + buffers + `,"bufferViews":[{"buffer":0,"byteLength":4}],"data":{"$Serializable":-1}}`,
		"fractional index":        `{` + buffers + `,"bufferViews":[{"buffer":0,"byteLength":4}],"data":{"$Serializable":0.5}}`,
		"index past views":        `{` + buffers + `,"bufferViews":[{"buffer":0,"byteLength":4}],"data":{"$Serializable":1}}`,
		"view past buffers":       `{` + buffers + `,"bufferViews":[{"buffer":1,"byteLength":4}],"data":{"$Serializable":0}}`,
		"view with negative ofst": `{` + buffers + `,"bufferViews":[{"buffer":0,"byteOffset":-1}],"data":{"$Serializable":0}}`,
	}

	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := jbtf.Unmarshal[embeddedBinaryStruct]([]byte(input))
			assert.Error(t, err)
		})
	}
}

func FuzzUnmarshalNeverPanics(f *testing.F) {
	valid := []any{
		true,
		123,
		"bababa",
		basicStruct{A: 123, B: true, C: "yee haw"},
		embeddedBinaryStruct{
			Basic:        basicStruct{A: 123, B: true, C: "yee haw"},
			Serializable: &jbtfSerializableStruct{data: 12345},
		},
		multipleEmbeddedBinaryStruct{
			SerializableA: &jbtfSerializableStruct{data: 12345},
			SerializableB: &jbtfSerializableStruct{data: 67890},
		},
		nesstedEmbeddedBinaryStruct{
			Nest: embeddedBinaryStruct{Serializable: &jbtfSerializableStruct{data: 12345}},
		},
		embeddedBytes{Bytes: &jbtf.Bytes{Data: []byte{1, 2, 3}}},
	}
	for _, v := range valid {
		out, err := jbtf.Marshal(v)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(out)
	}
	f.Add([]byte(`{"Buffers":[{}]}`))
	f.Add([]byte(`{"buffers":[{"uri":"data:application/octet-stream;base64,!!!"}]}`))
	f.Add([]byte(`{"buffers":[{"uri":"data:application/octet-stream;base64,OTAAAA=="}],"bufferViews":[{"buffer":3}],"data":{"$Serializable":0}}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		jbtf.Unmarshal[any](data)
		jbtf.Unmarshal[basicStruct](data)
		jbtf.Unmarshal[embeddedBinaryStruct](data)
		jbtf.Unmarshal[multipleEmbeddedBinaryStruct](data)
		jbtf.Unmarshal[nesstedEmbeddedBinaryStruct](data)
		jbtf.Unmarshal[embeddedBytes](data)
		jbtf.Unmarshal[embeddedPng](data)

		decoder, err := jbtf.NewDecoder(data)
		if err != nil {
			return
		}
		jbtf.Decode[embeddedBinaryStruct](decoder, data)
		jbtf.Decode[nesstedEmbeddedBinaryStruct](decoder, data)
	})
}

type embeddedBytes struct {
	Bytes *jbtf.Bytes
}

type embeddedPng struct {
	Image *jbtf.Png
}
