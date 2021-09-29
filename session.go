package xproject

import (
	"errors"
	"sync/atomic"
)

var ErrSessionClosed = errors.New("session closed")
var ErrSessionBlocked = errors.New("session blocked")

const sendChanSize int = 64

var newSessionId uint64

type Session struct {
	id           uint64
	codec        Codec
	eventHandler SessionEventHandler

	sendChan chan interface{}

	closed    int32
	closeChan chan int
}

type SessionEventHandler interface {
	// called when session received a message
	HandleMsg(session *Session, msg interface{})
	// called when session closed
	OnClose(session *Session)
}

func newSession(codec Codec, eventHandler SessionEventHandler) *Session {
	session := &Session{
		id:           atomic.AddUint64(&newSessionId, 1),
		codec:        codec,
		eventHandler: eventHandler,
		sendChan:     make(chan interface{}, sendChanSize),

		closeChan: make(chan int),
	}

	go session.sendLoop()
	go session.receiveLoop()

	return session
}

func (_this *Session) ID() uint64 {
	return _this.id
}

func (_this *Session) Codec() Codec {
	return _this.codec
}

func (_this *Session) IsClosed() bool {
	return atomic.LoadInt32(&_this.closed) == 1
}

func (_this *Session) Close() error {
	if atomic.CompareAndSwapInt32(&_this.closed, 0, 1) {
		close(_this.closeChan)

		if _this.sendChan != nil {
			close(_this.sendChan)
		}

		err := _this.codec.Close()

		_this.eventHandler.OnClose(_this)
		return err
	}

	return ErrSessionClosed
}

func (_this *Session) sendLoop() {
	for {
		select {
		case msg, ok := <-_this.sendChan:
			if !ok { // send channel closed already
				return
			} else if _this.codec.Send(msg) != nil {
				_this.Close()
				return
			}
		case <-_this.closeChan:
			return
		}
	}
}

func (_this *Session) receiveLoop() {
	for {
		msg, err := _this.codec.Receive()
		if err != nil {
			_this.Close()
			return
		}

		_this.eventHandler.HandleMsg(_this, msg)

		select {
		case <-_this.closeChan:
			return
		default:
			continue
		}
	}
}

func (_this *Session) Send(msg interface{}) error {
	if _this.IsClosed() {
		return ErrSessionClosed
	}

	select {
	case _this.sendChan <- msg:
		return nil
	default:
		_this.Close() // 发送阻塞就关闭，有更好的处理方法？
		return ErrSessionBlocked
	}
}
