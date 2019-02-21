package handlers

import (
	"fmt"
	"net/http"
	"github.com/fabiocampos/go-and-destroy/models"
	"github.com/fabiocampos/go-and-destroy/services"
	"github.com/gorilla/websocket"
)

func GameHandler(service *service.GameService) http.Handler {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			fmt.Println("WebsocketError: %v", err)
		}
		for {
			// Read the client's message
			_, msg, err := conn.ReadMessage()
			if err != nil {
				service.RemovePlayer(conn)
				conn.Close()
				return
			}
			service.ProcessAction(string(msg), conn.RemoteAddr().String())
		}
	})
}
