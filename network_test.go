package xproject

import (
	"log"
	"testing"
	"time"
)

type testAcceptorEventHandler struct {
}

func (_this *testAcceptorEventHandler) OnNewSession(session *Session) {
	log.Println("[Acceptor] accept a new session: ", session.ID())
}

func (_this *testAcceptorEventHandler) OnHandleSessionMsg(session *Session, msg interface{}) {
	log.Println("[Acceptor] receive a message from session: ", session.ID())
	if cmd, ok := msg.(string); ok {
		log.Println("cmd: ", cmd)

		session.Send(cmd)
	}
}

type testConnectorEventHandler struct {
}

func (_this *testConnectorEventHandler) OnHandleSessionMsg(session *Session, msg interface{}) {
	log.Println("[Connector] receive a message from session: ", session.ID())
	if cmd, ok := msg.(string); ok {
		log.Println("cmd: ", cmd)
	}
}

func TestNetwork(t *testing.T) {
	protocol := NewTextLineProtocol(16, 16)
	aEvHdler := &testAcceptorEventHandler{}
	cEvHdler := &testConnectorEventHandler{}

	acceptor, err := NewAcceptor("tcp", ":7777", protocol, aEvHdler)
	if err != nil {
		t.Fatal(err)
	}

	go acceptor.Run()

	connector := NewConnector("tcp", "127.0.0.1:7777", protocol, cEvHdler)

	connector.Run()

	<-time.After(time.Second * 2) // wait for connecting

	err = connector.Send("test")
	if err != nil {
		t.Fatal(err)
	}

	err = connector.Send("help")
	if err != nil {
		t.Fatal(err)
	}

	WaitForSystemExit()
}
