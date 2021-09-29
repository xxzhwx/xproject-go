package xproject

import (
	"errors"
	"log"
	"net"
	"sync/atomic"
	"time"
)

var ErrConnectorStopped = errors.New("connector stopped")
var ErrConnectorLost = errors.New("connector lost") // reconnecting

type Connector struct {
	network      string
	address      string
	protocol     Protocol
	eventHandler ConnectorEventHandler

	stopped  int32
	stopChan chan int

	session *Session
}

type ConnectorEventHandler interface {
	// called when session received a message
	OnHandleSessionMsg(session *Session, msg interface{})
}

func NewConnector(network string, address string, protocol Protocol, eventHandler ConnectorEventHandler) *Connector {
	return &Connector{
		network:      network,
		address:      address,
		protocol:     protocol,
		eventHandler: eventHandler,

		stopped:  1,
		stopChan: make(chan int),

		session: nil,
	}
}

// Implement SessionEventHandler.HandleMsg
func (_this *Connector) HandleMsg(session *Session, msg interface{}) {
	_this.eventHandler.OnHandleSessionMsg(session, msg)
}

// Implement SessionEventHandler.OnClose
func (_this *Connector) OnClose(session *Session) {
	_this.session = nil
}

func (_this *Connector) Run() {
	if atomic.CompareAndSwapInt32(&_this.stopped, 1, 0) {
		go _this.checkReconn()
	}
}

func (_this *Connector) Stop() {
	if atomic.CompareAndSwapInt32(&_this.stopped, 0, 1) {
		close(_this.stopChan)

		_this.session.Close()
	}
}

func (_this *Connector) Send(msg interface{}) error {
	if _this.IsStopped() {
		return ErrConnectorStopped
	}

	if _this.session == nil { //fixme thread-unsafe
		return ErrConnectorLost
	}

	return _this.session.Send(msg)
}

func (_this *Connector) IsStopped() bool {
	return atomic.LoadInt32(&_this.stopped) == 1
}

func (_this *Connector) checkReconn() {
	for {
		select {
		case <-_this.stopChan:
			return
		case <-time.After(time.Microsecond * 10):
			if !_this.IsStopped() && _this.session == nil {
				conn, err := net.DialTimeout(_this.network, _this.address, time.Second*2)
				if err != nil {
					// log.Println("dial time out")
					continue
				}

				codec, err := _this.protocol.NewCodec(conn)
				if err != nil {
					//todo log it
					log.Println("connect fail")
					continue
				}

				log.Println("connect success")
				_this.session = newSession(codec, _this)
			}
		}
	}
}
