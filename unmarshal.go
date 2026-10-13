package jbtf

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"reflect"
	"strings"
)

func jsonNameTag(v reflect.StructField) string {
	name := v.Name
	if altName, ok := v.Tag.Lookup("json"); ok {
		elements := strings.Split(altName, ",")
		name = elements[0]
	}
	return name
}

func jsonOmitTag(v reflect.StructField) bool {
	if altName, ok := v.Tag.Lookup("json"); ok {
		elements := strings.Split(altName, ",")
		if len(elements) < 2 {
			return false
		}
		if len(elements) != 2 {
			panic(fmt.Errorf("unimplemented json tag situation: %s", altName))
		}
		return elements[1] == "omitempty"
	}
	return false
}

func IsNilish(val any) bool {
	if val == nil {
		return true
	}

	v := reflect.ValueOf(val)
	k := v.Kind()
	switch k {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer,
		reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return v.IsNil()
	}

	return false
}

func BuildBuffer(buf Buffer) *bytes.Reader {
	r, err := buildBuffer(buf)
	if err != nil {
		panic(err)
	}
	return r
}

func buildBuffer(buf Buffer) (*bytes.Reader, error) {
	const dataUriIdentifier = "data:application/octet-stream;base64,"
	if strings.HasPrefix(buf.URI, dataUriIdentifier) {
		dataString := buf.URI[len(dataUriIdentifier):]
		data, err := base64.StdEncoding.DecodeString(dataString)
		if err != nil {
			return nil, fmt.Errorf("decoding buffer data uri: %w", err)
		}
		return bytes.NewReader(data), nil
	}
	return nil, fmt.Errorf("don't know how to interpret buffer: %+v", buf)
}

func rebuildBuffers(bufs []Buffer) ([]*bytes.Reader, error) {
	vals := make([]*bytes.Reader, len(bufs))
	for i, v := range bufs {
		r, err := buildBuffer(v)
		if err != nil {
			return nil, fmt.Errorf("buffer %d: %w", i, err)
		}
		vals[i] = r
	}
	return vals, nil
}

func interpretBufferView(value reflect.Value, rawData map[string]any, bufferViews []BufferView, buffers []*bytes.Reader) error {

	for value.Kind() == reflect.Pointer {
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return nil
	}

	valueType := value.Type()
	for fieldIndex := 0; fieldIndex < valueType.NumField(); fieldIndex++ {
		viewFieldValue := value.Field(fieldIndex)
		structField := valueType.Field(fieldIndex)

		if viewFieldValue.CanInterface() {
			i := viewFieldValue.Interface()
			_, isSerializable := i.(Serializable)
			if !isSerializable {
				if isStruct(i) {
					ahh, ok := rawData[structField.Name].(map[string]any)
					if ok {
						err := interpretBufferView(viewFieldValue, ahh, bufferViews, buffers)
						if err != nil {
							return err
						}
					}
				}
				continue
			}

			bufIndexKey := "$" + structField.Name
			bufIndexRaw, ok := rawData[bufIndexKey]

			// If this key isn't present, that means that field was nil during
			// serialization
			if !ok {
				continue
			}

			bufIndex, isNumber := bufIndexRaw.(float64)
			if !isNumber {
				return fmt.Errorf("buffer index '%s' was not a number", bufIndexKey)
			}

			if bufIndex < 0 || bufIndex >= float64(len(bufferViews)) || bufIndex != math.Trunc(bufIndex) {
				return fmt.Errorf("buffer index '%s' is invalid: %v", bufIndexKey, bufIndex)
			}
			bufView := bufferViews[int(bufIndex)]

			if bufView.Buffer < 0 || bufView.Buffer >= len(buffers) {
				return fmt.Errorf("buffer view %d references out of range buffer %d", int(bufIndex), bufView.Buffer)
			}
			buf := buffers[bufView.Buffer]
			_, err := buf.Seek(int64(bufView.ByteOffset), io.SeekStart)
			if err != nil {
				return err
			}

			typeEle := viewFieldValue.Type()
			for typeEle.Kind() == reflect.Pointer {
				typeEle = typeEle.Elem()
			}

			refNew := reflect.New(typeEle)
			instantiatedType := refNew.Interface()
			serializable, isSerializable := instantiatedType.(Serializable)
			if !isSerializable {
				return fmt.Errorf("field '%s' type %s is not serializable", structField.Name, typeEle)
			}

			err = serializable.Deserialize(buf)
			if err != nil {
				return err
			}

			if !viewFieldValue.CanSet() {
				return fmt.Errorf("field '%s' was found but can not be set", structField.Name)
			}

			viewFieldValue.Set(reflect.ValueOf(serializable))
			continue
		}
	}

	return nil
}

func ParseJsonUsingBuffers[T any](buffers []Buffer, bufferViews []BufferView, data []byte) (T, error) {
	var thingToParse T
	allBuffers, err := rebuildBuffers(buffers)
	if err != nil {
		return thingToParse, err
	}

	err = json.Unmarshal(data, &thingToParse)
	if err != nil {
		return thingToParse, err
	}

	rawData := make(map[string]any)
	err = json.Unmarshal(data, &rawData)
	if err != nil {
		return thingToParse, err
	}

	return thingToParse, interpretBufferView(reflect.ValueOf(&thingToParse), rawData, bufferViews, allBuffers)
}

func Unmarshal[T any](data []byte) (T, error) {
	format := &Schema[T]{}
	err := json.Unmarshal(data, format)
	if err != nil {
		return format.Data, err
	}

	rawData := make(map[string]any)
	err = json.Unmarshal(data, &rawData)
	if err != nil {
		return format.Data, err
	}

	specificData, ok := rawData["data"].(map[string]any)
	if ok {
		rawData = specificData
	}

	allBuffers, err := rebuildBuffers(format.Buffers)
	if err != nil {
		return format.Data, err
	}

	return format.Data, interpretBufferView(reflect.ValueOf(&format.Data), specificData, format.BufferViews, allBuffers)
}
