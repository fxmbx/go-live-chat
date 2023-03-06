package ws

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type Connection struct {
	Send   chan []byte
	socket *websocket.Conn
}

type Subscription struct {
	ID         string
	Connection Connection
}

type ClientManager struct {
	Clients    map[string]*Subscription
	Register   chan *Subscription
	Unregister chan *Subscription
	Broadcast  chan []byte
}

type Messagae struct {
	Sender    string `json:"sender,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	Content   string `json:"content,omitempty"`
}

var Manager = ClientManager{
	Clients:    make(map[string]*Subscription),
	Register:   make(chan *Subscription),
	Unregister: make(chan *Subscription),
	Broadcast:  make(chan []byte),
}

func (manager *ClientManager) Start() {
	for {
		log.Println("<---- pipeline communication ----->")
		select {
		case conn := <-Manager.Register:
			log.Println("new user joined in %v", conn.ID)
			Manager.Clients[conn.ID] = conn
			jsonMessage, _ := json.Marshal(&Messagae{Content: "successful connection to socket service"})
			conn.Connection.Send <- jsonMessage
		case conn := <-Manager.Unregister:
			log.Printf("user %s left ", conn.ID)
			if _, ok := Manager.Clients[conn.ID]; ok {
				jsonMessage, _ := json.Marshal(&Messagae{Content: "a socket has been disconnected"})
				conn.Connection.Send <- jsonMessage
				close(conn.Connection.Send)
				delete(Manager.Clients, conn.ID)
			}
		case message := <-Manager.Broadcast:
			MessageStruct := Messagae{}
			json.Unmarshal(message, &MessageStruct)
			for id, conn := range Manager.Clients {
				if id != createId(MessageStruct.Recipient, MessageStruct.Sender) {
					continue
				}
				select {
				case conn.Connection.Send <- message:
				default:
					close(conn.Connection.Send)
					delete(Manager.Clients, conn.ID)
				}
			}
		}
	}
}

func createId(uid, touid string) string {
	return uid + "_" + touid
}
