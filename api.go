package xproject

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type Protocol interface {
	NewCodec(rw io.ReadWriter) (Codec, error)
}

type ProtocolFunc func(rw io.ReadWriter) (Codec, error)

func (f ProtocolFunc) NewCodec(rw io.ReadWriter) (Codec, error) {
	return f(rw)
}

// Codec should be thread-safe
type Codec interface {
	Receive() (interface{}, error)
	Send(interface{}) error
	Close() error
}

// system exit
var sysExitChan = make(chan os.Signal, 1)

func WaitForSystemExit() {
	signal.Notify(sysExitChan, os.Interrupt, syscall.SIGTERM)
	<-sysExitChan

	log.Println("System Exit ...")
}
