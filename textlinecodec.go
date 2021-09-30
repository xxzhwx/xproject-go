package xproject

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

type TextLineProtocol struct {
	rBufSz int
	wBufSz int
}

func NewTextLineProtocol(readBufSize int, writeBufSize int) *TextLineProtocol {
	return &TextLineProtocol{
		rBufSz: readBufSize,
		wBufSz: writeBufSize,
	}
}

func (_this *TextLineProtocol) NewCodec(rw io.ReadWriter) (Codec, error) {
	codec := &TextLineCodec{
		r: bufio.NewReaderSize(rw, _this.rBufSz),
		w: bufio.NewWriterSize(rw, _this.wBufSz),
	}

	codec.c, _ = rw.(io.Closer)

	return codec, nil
}

type TextLineCodec struct {
	r *bufio.Reader
	w *bufio.Writer

	c io.Closer
}

func (_this *TextLineCodec) Receive() (interface{}, error) {
	str, err := _this.r.ReadString('\n')
	// log.Printf("Receive: '%s'\n", str)

	str = strings.TrimSuffix(str, "\n")
	return str, err
}

func (_this *TextLineCodec) Send(msg interface{}) error {
	str, ok := msg.(string)
	if !ok {
		return errors.New("not text")
	}

	if !strings.HasSuffix(str, "\n") {
		str += "\n"
	}

	// log.Printf("Send: '%s'\n", str)
	_, err := _this.w.WriteString(str)
	_this.w.Flush()
	return err
}

func (_this *TextLineCodec) Close() error {
	if _this.c != nil {
		_this.c.Close()
	}
	return nil
}
