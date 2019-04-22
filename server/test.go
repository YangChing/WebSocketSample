package main

import (
    "fmt"
    "github.com/satori/go.uuid"
    "github.com/gorilla/websocket"
)

type Client struct {
    id     string
    socket *websocket.Conn
    send   chan []byte
}

func main() {
    u1, _ := uuid.NewV4()
    fmt.Printf(u1.String())

    //client := &Client{id: uuid.NewV4().String(), socket: conn, send: make(chan []byte)}
    //fmt.Println(client)
}
