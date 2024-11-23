package drive

import (
	"errors"
	"golang.org/x/sys/unix"
	"sync"
)

type Poll struct {
	mu    sync.Mutex // 保护内部数据的锁
	fds   []unix.PollFd
	fdMap map[int]int // 用于快速查找文件描述符在 fds 中的索引
}

func NewPoll() Poller {
	return &Poll{
		fds:   make([]unix.PollFd, 0),
		fdMap: make(map[int]int),
		mu:    sync.Mutex{},
	}
}

func (p *Poll) AddRead(fd int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.fdMap[fd]; exists {
		return errors.New("file descriptor already exists")
	}

	p.fds = append(p.fds, unix.PollFd{
		Fd:     int32(fd),
		Events: unix.POLLIN,
	})
	p.fdMap[fd] = len(p.fds) - 1
	return nil
}

func (p *Poll) AddWrite(fd int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.fdMap[fd]; exists {
		return errors.New("file descriptor already exists")
	}

	p.fds = append(p.fds, unix.PollFd{
		Fd:     int32(fd),
		Events: unix.POLLOUT,
	})
	p.fdMap[fd] = len(p.fds) - 1
	return nil
}

func (p *Poll) ModRead(fd int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	index, exists := p.fdMap[fd]
	if !exists {
		return errors.New("file descriptor does not exist")
	}

	p.fds[index].Events = unix.POLLIN
	return nil
}

func (p *Poll) ModWrite(fd int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	index, exists := p.fdMap[fd]
	if !exists {
		return errors.New("file descriptor does not exist")
	}

	p.fds[index].Events = unix.POLLOUT
	return nil
}

func (p *Poll) Remove(fd int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	index, exists := p.fdMap[fd]
	if !exists {
		return errors.New("file descriptor does not exist")
	}

	// 移除 fd
	p.fds = append(p.fds[:index], p.fds[index+1:]...)
	delete(p.fdMap, fd)

	// 更新 fdMap 的索引
	for i := index; i < len(p.fds); i++ {
		p.fdMap[int(p.fds[i].Fd)] = i
	}

	return nil
}

// Polling .
func (p *Poll) Polling() ([]ExternalEvent, error) {

	n, err := unix.Poll(p.fds, -1)
	if err != nil {
		return nil, err
	}

	external := make([]ExternalEvent, 0, n)

	for _, fd := range p.fds {

		// 可读
		if fd.Revents&unix.POLLIN != 0 {
			external = append(external, ExternalEvent{
				Fd:     int(fd.Fd),
				Opcode: OpcodeRead,
			})
		}

		// 可写
		if fd.Revents&unix.POLLOUT != 0 {
			external = append(external, ExternalEvent{
				Fd:     int(fd.Fd),
				Opcode: OpcodeWrite,
			})
		}
	}

	return external, nil
}
