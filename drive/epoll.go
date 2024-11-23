package drive

import (
	"log"
	"syscall"
)

type epoll struct {
	fd     int
	events []syscall.EpollEvent
}

func EPoll() Poller {
	fd, err := syscall.EpollCreate1(syscall.EPOLL_CLOEXEC)
	if err != nil {
		log.Panicln(err)
	}
	return &epoll{
		fd:     fd,
		events: make([]syscall.EpollEvent, 128),
	}
}

// Polling .
func (e *epoll) Polling() ([]ExternalEvent, error) {

	n, err := syscall.EpollWait(e.fd, e.events, -1)
	if err != nil {
		return nil, err
	}

	external := make([]ExternalEvent, 0)
	for i := 0; i < n; i++ {
		event := e.events[i]
		opcode := OpcodeRead

		// 可写
		if event.Events&syscall.EPOLLOUT == syscall.EPOLLOUT {
			opcode = OpcodeWrite
		}

		external = append(external, ExternalEvent{
			Fd:     int(event.Fd),
			Opcode: opcode,
		})
	}

	return external, nil
}

func (e *epoll) AddRead(fd int) error {
	return syscall.EpollCtl(e.fd, syscall.EPOLL_CTL_ADD, fd, &syscall.EpollEvent{
		Fd:     int32(fd),
		Events: syscall.EPOLLIN | syscall.EPOLLPRI,
	})
}

func (e *epoll) AddWrite(fd int) error {
	return syscall.EpollCtl(e.fd, syscall.EPOLL_CTL_ADD, fd, &syscall.EpollEvent{
		Events: syscall.EPOLLOUT,
		Fd:     int32(fd),
	})
}

func (e *epoll) ModRead(fd int) error {
	return syscall.EpollCtl(e.fd, syscall.EPOLL_CTL_MOD, fd, &syscall.EpollEvent{
		Events: syscall.EPOLLIN | syscall.EPOLLPRI,
		Fd:     int32(fd),
	})
}

func (e *epoll) ModWrite(fd int) error {
	return syscall.EpollCtl(e.fd, syscall.EPOLL_CTL_MOD, fd, &syscall.EpollEvent{
		Events: syscall.EPOLLOUT,
		Fd:     int32(fd),
	})
}

// Remove .
func (e *epoll) Remove(fd int) error {
	return syscall.EpollCtl(e.fd, syscall.EPOLL_CTL_DEL, fd, nil)
}
