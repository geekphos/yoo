package socket_client

import (
	"github.com/gorilla/websocket"
	"sync"
)

var (
	conns = make(map[string]*websocket.Conn)
	mt    = sync.RWMutex{}
)

func AddConn(key string, conn *websocket.Conn) {
	mt.Lock()
	defer mt.Unlock()

	conns[key] = conn
}

func GetConn(key string) *websocket.Conn {
	mt.RLock()
	defer mt.RUnlock()

	return conns[key]
}

func RemoveConn(key string) {
	mt.Lock()
	defer mt.Unlock()

	delete(conns, key)
}
