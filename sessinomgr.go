package xproject

import "sync"

const sessionMapNum = 32

type SessionMgr struct {
	sessionMaps [sessionMapNum]sessionMap
	disposeOnce sync.Once
	disposeWait sync.WaitGroup
}

type sessionMap struct {
	sync.RWMutex
	sessions map[uint64]*Session
	disposed bool
}

func NewSessionMgr() *SessionMgr {
	mgr := &SessionMgr{}
	for i := 0; i < len(mgr.sessionMaps); i++ {
		mgr.sessionMaps[i].sessions = make(map[uint64]*Session)
	}
	return mgr
}

func (_this *SessionMgr) Dispose() {
	_this.disposeOnce.Do(func() {
		for i := 0; i < len(_this.sessionMaps); i++ {
			smap := &_this.sessionMaps[i]
			smap.Lock()
			smap.disposed = true
			for _, session := range smap.sessions {
				session.Close()
			}
			smap.Unlock()
		}
		_this.disposeWait.Wait()
	})
}

func (_this *SessionMgr) NewSession(codec Codec) *Session {
	session := newSession(_this, codec)
	_this.putSession(session)
	return session
}

func (_this *SessionMgr) GetSession(sessionID uint64) *Session {
	smap := &_this.sessionMaps[sessionID%sessionMapNum]

	smap.RLock()
	defer smap.RUnlock()

	session := smap.sessions[sessionID]
	return session
}

func (_this *SessionMgr) putSession(session *Session) {
	smap := &_this.sessionMaps[session.id%sessionMapNum]

	smap.Lock()
	defer smap.Unlock()

	if smap.disposed {
		session.Close()
		return
	}

	smap.sessions[session.id] = session
	_this.disposeWait.Add(1)
}

func (_this *SessionMgr) delSession(session *Session) {
	smap := &_this.sessionMaps[session.id%sessionMapNum]

	smap.Lock()
	defer smap.Unlock()

	if _, ok := smap.sessions[session.id]; ok {
		delete(smap.sessions, session.id)
		_this.disposeWait.Done()
	}
}
