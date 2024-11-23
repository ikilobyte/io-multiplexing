package main

import (
	"bytes"
	"fmt"
	"io"
	"io-multiplexing/drive"
	"log"
	"os"
	"syscall"
)

type OnConnect func(client *Client)
type OnMessage func(client *Client, message []byte, n int)
type OnClose func(client *Client)

type Server struct {
	fd        int
	epfd      int
	poller    drive.Poller
	onConnect OnConnect
	onMessage OnMessage
	onClose   OnClose
}

func NewServer(poller drive.Poller, port int) *Server {

	fd, err := syscall.Socket(
		syscall.AF_INET,
		syscall.SOCK_STREAM|syscall.SOCK_CLOEXEC,
		syscall.IPPROTO_TCP,
	)
	if err != nil {
		log.Panicln(err)
	}

	addr := syscall.SockaddrInet4{
		Addr: [4]byte{0, 0, 0, 0},
		Port: port,
	}
	if err := syscall.Bind(fd, &addr); err != nil {
		log.Panicln(err)
	}

	// 非阻塞
	if err := syscall.SetNonblock(fd, true); err != nil {
		log.Panicln(err)
	}

	_ = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	_ = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEPORT, 1)

	// 监听
	if err := syscall.Listen(fd, syscall.SOMAXCONN); err != nil {
		log.Panicln(err)
	}

	fmt.Printf("[%d]tcp server listen at 0.0.0.0:%d\n", os.Getpid(), port)
	if err := poller.AddRead(fd); err != nil {
		log.Panicln(err)
	}

	// 创建server
	return &Server{
		fd:     fd,
		poller: poller,
	}
}

// Serve .
func (s *Server) Serve() error {
	for {

		events, err := s.poller.Polling()
		if err != nil {
			fmt.Printf("polling error: %v\n", err)
			continue
		}

		for _, event := range events {

			// server 事件
			if event.Fd == s.fd {
				client, err := s.accept()
				if err != nil {
					fmt.Printf("accept error: %v\n", err)
					continue
				}
				clients[event.Fd] = client
				s.onConnect(client)
				continue
			}

			client, ok := clients[event.Fd]
			if !ok {
				fmt.Printf("client %d not found\n", event.Fd)
				continue
			}

			// 可读
			if event.Opcode == drive.OpcodeRead {
				n, buf, err := client.Read(1024)
				if err != nil {
					if err == io.EOF {
						client.Close()
						s.onClose(client)
						continue
					}
					fmt.Printf("read error: %v\n", err)
					continue
				}

				// 断开连接
				if n == 0 {
					client.Close()
					s.onClose(client)
					continue
				}

				s.onMessage(client, buf, n)
			}

			// 可写
			if event.Opcode == drive.OpcodeWrite {
				client.writing()
			}
		}
	}
}

// accept .
func (s *Server) accept() (*Client, error) {
	fd, sockaddr, err := syscall.Accept(s.fd)
	if err != nil {
		return nil, err
	}

	client := &Client{
		fd:     fd,
		addr:   sockaddr,
		buffer: &bytes.Buffer{},
		poller: s.poller,
	}
	clients[client.fd] = client

	// 设置为非阻塞
	if err := syscall.SetNonblock(fd, true); err != nil {
		return nil, err
	}

	_ = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_SNDBUF, 1024*128)

	// 添加可读事件
	if err = s.poller.AddRead(fd); err != nil {
		return nil, err
	}
	return client, nil
}
