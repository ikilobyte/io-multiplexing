package main

import (
	"fmt"
	"io-multiplexing/drive"
	"log"
	"strings"
)

var clients map[int]*Client

func init() {
	clients = make(map[int]*Client)
}

func main() {

	//poller := drive.NewSelect()
	//poller := drive.NewPoll()
	poller := drive.NewEPoll()
	server := NewServer(poller, 7000)

	server.onConnect = func(client *Client) {
		fmt.Println("onConnect", client.fd)
		client.Write([]byte(strings.Repeat("A", 1024*1024*10)))
		client.Write([]byte("hello world"))
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
