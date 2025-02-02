package encodings_test

import (
	"testing"

	"github.com/titosilva/drmchain-pos/network/encodings"
)

type msgType struct {
	Seq     int
	Cmd     string
	Payload []byte
}

func Test__Decode__ShouldReturnObject__EncodedWithEncode(t *testing.T) {
	encoded, err := encodings.Encode(msgType{
		Seq:     1,
		Cmd:     "cmd",
		Payload: []byte("payload"),
	})

	if err != nil {
		t.Error(err)
	}

	decoded := msgType{}
	err = encodings.Decode(encoded, &decoded)

	if err != nil {
		t.Error(err)
	}

	if decoded.Seq != 1 {
		t.Error("Expected seq to be 1")
	}

	if decoded.Cmd != "cmd" {
		t.Error("Expected cmd to be 'cmd'")
	}

	if string(decoded.Payload) != "payload" {
		t.Error("Expected payload to be 'payload'")
	}
}
