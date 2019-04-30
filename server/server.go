package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

// ClientManager is used to manager client
type ClientManager struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

// Client
type Client struct {
	socket   *websocket.Conn
	send     chan []byte
	username string
}

// Message
type Message struct {
	Sender    string `json:"sender,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	Content   string `json:"content,omitempty"`
	Username  string `json:"username,omitempty"`
}

// Post
type Post struct {
	Username string `json:"username,omitempty"`
	Message  string `json:"message"`
	Time     string `json:"time"`
}

var manager = ClientManager{
	broadcast:  make(chan []byte),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	clients:    make(map[*Client]bool),
}

func (manager *ClientManager) start() {
	for {
		select {
		case conn := <-manager.register:
			manager.clients[conn] = true
			jsonMessage, _ := json.Marshal(&Post{Username: conn.username, Time: time.Now().Format("2006-01-02 15:04:05"), Message: "entry room"})
			manager.send(jsonMessage, conn)
		case conn := <-manager.unregister:
			if _, ok := manager.clients[conn]; ok {
				close(conn.send)
				delete(manager.clients, conn)
				jsonMessage, _ := json.Marshal(&Post{Username: conn.username, Time: time.Now().Format("2006-01-02 15:04:05"), Message: "leave room"})
				manager.send(jsonMessage, conn)
			}
		case message := <-manager.broadcast:
			for conn := range manager.clients {
				select {
				case conn.send <- message:
				default:
					close(conn.send)
					delete(manager.clients, conn)
				}
			}
		}
	}
}

func (manager *ClientManager) send(message []byte, ignore *Client) {
	for conn := range manager.clients {
		if conn != ignore {
			conn.send <- message
		}
	}
}

func (c *Client) read() {
	defer func() {
		manager.unregister <- c
		c.socket.Close()
	}()

	for {
		var p Post
		err := c.socket.ReadJSON(&p)
		if err != nil {
			manager.unregister <- c
			c.socket.Close()
			fmt.Println("err", err)
			break
		}
		jsonMessage, _ := json.Marshal(&Post{Username: p.Username, Message: p.Message, Time: p.Time})
		manager.broadcast <- jsonMessage
	}
}

func (c *Client) write() {
	defer func() {
		c.socket.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.socket.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("ip:")
	scanner.Scan()
	ip := scanner.Text()
	if ip == "" {
		ip = "127.0.0.1:12345"
	}
	fmt.Println("Starting application...")
	fmt.Println(ip)
	go manager.start()
	http.HandleFunc("/ws", wsPage)
	err := http.ListenAndServe(ip, nil)
	if err != nil {
		fmt.Println(err)
	}

}

func wsPage(res http.ResponseWriter, req *http.Request) {
	conn, error := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(res, req, nil)
	if error != nil {
		http.NotFound(res, req)
		return
	}

	client := &Client{socket: conn, send: make(chan []byte), username: req.Header.Get("username")}
	manager.register <- client

	go client.read()
	go client.write()
}
