package main

import (
	"fmt"
	"io-multiplexing/drive"
	"log"
)

var clients map[int]*Client

func init() {
	clients = make(map[int]*Client)
}

func main() {

	poller := drive.EPoll()
	server := NewServer(poller, 7000)

	server.onConnect = func(client *Client) {
		fmt.Println("onConnect", client.fd)
	}

	server.onMessage = func(client *Client, message []byte, n int) {
		fmt.Printf("onMessage %s", string(message[:n]))
		client.Write([]byte(fmt.Sprintf("reply %s", string(message[:n]))))
	}

	server.onClose = func(client *Client) {
		fmt.Println("onClose", client.fd)
	}

	if err := server.Serve(); err != nil {
		log.Panicln(err)
	}
}
