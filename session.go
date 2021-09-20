package main

import "time"

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

func (sessions *sessions) cleanSessions() {
	for {
		for k, v := range sessions.sessions {
			if time.Now().After(v.expires) {
				delete(sessions.sessions, k)
			}
		}
	}
}

func (sessions *sessions) AddSession(key string, value interface{}) {
	sessions.sessions[key] = sessionInfo{
		value:   value,
		expires: time.Now(),
	}
}

func (sessions *sessions) RemoveSession(key string) {
	delete(sessions.sessions, key)
}
