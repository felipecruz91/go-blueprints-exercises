package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

type room struct {
	// forward is a channel that holds incoming messages
	// that should be forward to other clients
	forward chan []byte
	// join is a channel for clients whishing to join the room.
	join chan *client
	// leave is a channel for clients wishing to leave the room.
	leave chan *client
	// clients holds all current clients in this room.
	clients map[*client]bool
}

func newRoom() *room {
	return &room{
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
	}
}

func (r *room) run() {
	for { //loop forever
		select {
		case client := <-r.join:
			r.clients[client] = true
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.send)
		case msg := <-r.forward:
			// forward message to all clients
			for client := range r.clients {
				select {
				case client.send <- msg:
					// send the message
				default:
					// failed to send
					delete(r.clients, client)
					close(client.send)
				}
			}
		}
	}
}

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

// A room can act now as a handler
func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// upgrade the HTTP connection to use sockets
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServerHTTP:", err)
		return
	}
	// create our client
	client := &client{
		socket: socket,
		send:   make(chan []byte, messageBufferSize),
		room:   r,
	}
	r.join <- client
	// defer the leaving op for when the client is finished
	defer func() { r.leave <- client }()
	go client.write()
	// call the read method in the main thread, which will block
	// operations (keeping the connection alive) until it's
	// time to close it.
	client.read()
}
