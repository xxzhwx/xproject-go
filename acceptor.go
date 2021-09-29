package xproject

import (
	"net"
)

type Acceptor struct {
	listener     net.Listener
	protocol     Protocol
	eventHandler AcceptorEventHandler
	sessionMgr   *SessionMgr
}

type AcceptorEventHandler interface {
	// called when acceptor accepted a new session
	OnNewSession(session *Session)
	// called when session received a message
	OnHandleSessionMsg(session *Session, msg interface{})
	// called when session closed
	// OnCloseSession(session *Session)
}

func NewAcceptor(network string, address string, protocol Protocol, eventHandler AcceptorEventHandler) (*Acceptor, error) {
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}

	return &Acceptor{
		listener:     listener,
		protocol:     protocol,
		eventHandler: eventHandler,
		sessionMgr:   NewSessionMgr(),
	}, nil
}

// Implement SessionEventHandler.HandleMsg
func (_this *Acceptor) HandleMsg(session *Session, msg interface{}) {
	_this.eventHandler.OnHandleSessionMsg(session, msg)
}

// Implement SessionEventHandler.OnClose
func (_this *Acceptor) OnClose(session *Session) {
	_this.sessionMgr.DelSession(session) // New and Del both controlled by Acceptor
	// _this.eventHandler.OnCloseSession(session) // use goroutine ?
}

func (_this *Acceptor) Run() error {
	for {
		conn, err := _this.listener.Accept()
		if err != nil {
			return err
		}

		codec, err := _this.protocol.NewCodec(conn)
		if err != nil {
			conn.Close()
			return err
		}

		session := _this.sessionMgr.NewSession(codec, _this)
		go _this.eventHandler.OnNewSession(session)
	}
}

func (_this *Acceptor) Stop() {
	_this.listener.Close()
	_this.sessionMgr.Dispose()
}

func (_this *Acceptor) GetSession(sessionID uint64) *Session {
	return _this.sessionMgr.GetSession(sessionID)
}
