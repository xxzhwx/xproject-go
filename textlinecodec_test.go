package xproject

import (
	"bytes"
	"log"
	"testing"
)

func TestTextLineCodec(t *testing.T) {
	protocol := NewTextLineProtocol(16, 16)

	var stream bytes.Buffer
	codec, _ := protocol.NewCodec(&stream)

	err := codec.Send("help")
	if err != nil {
		t.Fatal(err)
	}

	recvMsg, err := codec.Receive()
	if err != nil {
		t.Fatal(err)
	}

	if str, ok := recvMsg.(string); ok {
		log.Println("recv: ", str)
	}
}
