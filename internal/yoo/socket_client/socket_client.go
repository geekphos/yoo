package socket_client

import (
	"github.com/gorilla/websocket"
	"sync"
)

var (
	conns = make(map[string]*websocket.Conn)
	mt    = sync.Mutex{}
)

func AddConn(key string, conn *websocket.Conn) {
	mt.Lock()
	defer mt.Unlock()

	conns[key] = conn
}

func RemoveConn(key string) {
	mt.Lock()
	defer mt.Unlock()

	delete(conns, key)
}

func WriteMessage(key, msg string) bool {
	mt.Lock()
	defer mt.Unlock()
	if conn, ok := conns[key]; ok {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
			return false
		}
		return true
	}

	return false
}

func WriteJSON(key string, v interface{}) bool {
	mt.Lock()
	defer mt.Unlock()
	if conn, ok := conns[key]; ok {
		if err := conn.WriteJSON(v); err != nil {
			return false
		}
		return true
	}

	return false
}
