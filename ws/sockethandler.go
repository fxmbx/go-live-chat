package ws

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func WsHandler(ctx *gin.Context) {
	uid := ctx.Query("uid")
	touid := ctx.Query("to_uid")
	fmt.Println(uid + "and to: " + touid)
	// conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	conn, err := (&websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}).Upgrade(ctx.Writer, ctx.Request, nil)

	if err != nil {
		http.NotFound(ctx.Writer, ctx.Request)
		return
	}
	clientSubscription := &Subscription{
		ID: createId(uid, touid),
		Connection: Connection{
			Send:   make(chan []byte),
			socket: conn,
		},
	}
	Manager.Register <- clientSubscription
	go clientSubscription.ReadPump()
	go clientSubscription.WritePump()
}

func (s *Subscription) ReadPump() {
	defer func() {
		// defer close(c.Connection.Send)
		s.Connection.socket.Close()
		Manager.Unregister <- s
	}()

	for {
		s.Connection.socket.PongHandler()
		_, message, err := s.Connection.socket.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			Manager.Unregister <- s
			s.Connection.socket.Close()
			break
		}
		log.Printf("message read client: %s", string(message))
		Manager.Broadcast <- message
	}
}

func (s *Subscription) WritePump() {
	defer func() {
		s.Connection.socket.Close()
	}()

	for {
		select {
		case message, ok := <-s.Connection.Send:
			if !ok {
				s.Connection.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			log.Printf("mesage sent to client: %s", string(message))
			s.Connection.socket.WriteMessage(websocket.TextMessage, message)
		}
	}
}
