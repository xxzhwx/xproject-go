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

		session := _this.sessionMgr.NewSession(codec)
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
