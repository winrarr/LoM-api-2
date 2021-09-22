package sessions

import (
	"fmt"
	"time"
)

type sessions struct {
	sessions map[string]sessionInfo
}

type sessionInfo struct {
	value   interface{}
	expires time.Time
}

func Sessions() sessions {
	sessions := sessions{
		sessions: make(map[string]sessionInfo),
	}

	go sessions.cleanSessions()

	return sessions
}

func (sessions *sessions) AddSession(key string, value interface{}) {
	sessions.sessions[key] = sessionInfo{
		value:   value,
		expires: time.Now().Add(time.Hour),
	}
}

func (sessions *sessions) RemoveSession(key string) {
	delete(sessions.sessions, key)
}

func (sessions *sessions) printSessions() {
	println("printing sessions")
	for k, v := range sessions.sessions {
		fmt.Println("key: " + k + ", value: " + v.expires.String())
	}
}

func (sessions *sessions) cleanSessions() {
	for {
		for k, v := range sessions.sessions {
			if time.Now().After(v.expires) {
				sessions.RemoveSession(k)
			}
		}
	}
}
