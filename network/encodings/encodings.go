package encodings

import (
	"encoding/asn1"
	"errors"
	"reflect"
)

// TODO: return to asn1

func Encode(data any) ([]byte, error) {
	if reflect.TypeOf(data).Kind() == reflect.Ptr {
		data = reflect.ValueOf(data).Elem().Interface()
	}

	return asn1.Marshal(data)
}

func Decode(data []byte, target any) error {
	rest, err := asn1.Unmarshal(data, target)

	if len(rest) > 0 {
		return errors.New("not all data was consumed")
	}

	return err
}

// func Encode(data any) ([]byte, error) {
// 	return json.Marshal(data)
// }

// func Decode(data []byte, target any) error {
// 	err := json.Unmarshal(data, target)

// 	return err
// }

func DecodeAs[T any](data []byte) (T, error) {
	var target T
	err := Decode(data, &target)

	return target, err
}
