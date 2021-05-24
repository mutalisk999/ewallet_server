package session

import (
	"github.com/kataras/iris/core/errors"
	"github.com/kataras/iris/websocket"
	"sync"
	"time"
	"utils"
)

type SessionValue struct {
	WebConn    websocket.Connection
	ServerId   int
	ServerName string
	PubKey     string
	Random     string
	IsLogin    bool
	CreateTime time.Time
	UpdateTime time.Time
}

type SessionMgr struct {
	//SessionStore map[websocket.Connection]SessionValue
	SessionStore map[string]SessionValue
	Mutex        *sync.Mutex
}

var GlobalSessionMgr *SessionMgr

func InitSessionMgr() {
	GlobalSessionMgr = new(SessionMgr)
	GlobalSessionMgr.InitSessionStore()
	GlobalSessionMgr.Mutex = new(sync.Mutex)
}

func (m *SessionMgr) InitSessionStore() {
	//m.SessionStore = make(map[websocket.Connection]SessionValue)
	m.SessionStore = make(map[string]SessionValue)
}

func (m SessionMgr) HasSessionId(webConn websocket.Connection) bool {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	_, ok := m.SessionStore[webConn.ID()]
	return ok
}

func (m SessionMgr) GetSessionValue(webConn websocket.Connection) (SessionValue, bool) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	sessionValue, ok := m.SessionStore[webConn.ID()]
	if !ok {
		return SessionValue{}, false
	}
	sessionValue.UpdateTime = time.Now()
	m.SessionStore[webConn.ID()] = sessionValue
	return sessionValue, true
}



func (m *SessionMgr) NewNullSessionValue(webConn websocket.Connection) (*SessionValue, error) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()
	var sessionValue SessionValue
	sessionValue.WebConn = webConn
	sessionValue.ServerId = -1
	sessionValue.IsLogin = false
	sessionValue.Random = utils.RandomString(10)
	_, ok := m.SessionStore[webConn.ID()]
	if !ok {
		sessionValue.CreateTime = time.Now()
		sessionValue.UpdateTime = time.Now()
		m.SessionStore[sessionValue.WebConn.ID()] = sessionValue
		return &sessionValue, nil
	}
	return nil, errors.New("already exist same connection!")
}

func (m *SessionMgr) NewSessionValue(sessionValue SessionValue) (bool, error) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	_, ok := m.SessionStore[sessionValue.WebConn.ID()]
	if !ok {
		sessionValue.CreateTime = time.Now()
		sessionValue.UpdateTime = time.Now()
		m.SessionStore[sessionValue.WebConn.ID()] = sessionValue
		return true, nil
	}
	return false, errors.New("already exist same connection!")
}


func (m *SessionMgr) UpdateSessionValue(sessionValue SessionValue) (bool, error) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	_, ok := m.SessionStore[sessionValue.WebConn.ID()]
	if ok {
		sessionValue.UpdateTime = time.Now()
		m.SessionStore[sessionValue.WebConn.ID()] = sessionValue
		return true, nil
	}
	return false, errors.New("not find connection!")
}


func (m *SessionMgr) DeleteSessionValue(webConn websocket.Connection) {
	m.Mutex.Lock()
	defer m.Mutex.Unlock()

	delete(m.SessionStore, webConn.ID())
}

func (m SessionMgr) IsLogin(webConn websocket.Connection) (bool, error) {
	sessionValue, ok := m.GetSessionValue(webConn)
	if !ok {
		return false, errors.New("session id not exist")
	}
	if sessionValue.ServerId > -1 {
		return true, nil
	}
	return false, nil
}