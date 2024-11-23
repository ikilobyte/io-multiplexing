package main

import (
	"bytes"
	"io"
	"io-multiplexing/drive"
	"syscall"
)

type Client struct {
	fd      int
	addr    syscall.Sockaddr
	buffer  *bytes.Buffer
	poller  drive.Poller
	written int
}

// writing .
func (c *Client) writing() {
	buffer := c.buffer.Bytes()[c.written:]
	size := len(buffer)
	n, err := syscall.Write(c.fd, buffer)
	if err != nil {
		if err == io.EOF {
			c.Close()
		}
		return
	}

	c.written += n
	if n == size {
		c.buffer.Reset()
		_ = c.poller.ModRead(c.fd)
		c.written = 0
	}
}

func (c *Client) Read(n int) (int, []byte, error) {
	buf := make([]byte, n)
	n, err := syscall.Read(c.fd, buf)
	if err != nil {
		return 0, nil, err
	}

	data := buf[:n]
	return n, data, nil
}

// Write .
func (c *Client) Write(data []byte) (int, error) {

	size := len(data)

	if c.buffer.Len() >= 1 {
		c.buffer.Write(data)
		return size, nil
	}

	n, err := syscall.Write(c.fd, data)
	if err != nil {
		if err == io.EOF {
			c.Close()
			return 0, err
		}
		return 0, err
	}

	if n <= 0 {
		c.Close()
		return 0, io.EOF
	}

	if n < size {
		c.written += n
		c.buffer.Write(data[n:])
		err = c.poller.ModWrite(c.fd)
	} else {
		c.buffer.Reset()
		_ = c.poller.ModRead(c.fd)
		c.written = 0
	}

	return size, nil
}

// Close 断开连接
func (c *Client) Close() {
	_ = syscall.Close(c.fd)
	_ = c.poller.Remove(c.fd)
	delete(clients, c.fd)
}
