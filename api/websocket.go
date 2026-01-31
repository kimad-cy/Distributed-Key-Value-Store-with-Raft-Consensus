package api

import (
	"net/http"
	"time"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, 
}

func (h *Handlers) StatusWebsocket(w http.ResponseWriter, r *http.Request) {
    conn, _ := upgrader.Upgrade(w, r, nil)
    defer conn.Close()

    ticker := time.NewTicker(250 * time.Millisecond)
    defer ticker.Stop()

    for range ticker.C {
        state := h.Node.GetFullStatus()
        state["now"] = time.Now().UnixMilli()
        if err := conn.WriteJSON(state); err != nil {
            break
        }
    }
}