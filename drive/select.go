package drive

import (
	"syscall"
)

type Select struct {
	read  syscall.FdSet
	write syscall.FdSet
	max   int
	fds   map[int]int
}

func NewSelect() Poller {
	return &Select{
		read:  syscall.FdSet{},
		write: syscall.FdSet{},
		max:   0,
		fds:   make(map[int]int),
	}
}

func (s *Select) AddRead(fd int) error {
	s.read.Bits[fd/64] |= 1 << (fd % 64)
	if fd > s.max {
		s.max = fd
	}
	s.fds[fd] = fd
	return nil
}

func (s *Select) AddWrite(fd int) error {
	s.write.Bits[fd/64] |= 1 << (fd % 64)
	if fd > s.max {
		s.max = fd
	}

	s.fds[fd] = fd
	return nil
}

func (s *Select) ModRead(fd int) error {
	s.write.Bits[fd/64] &^= 1 << (fd % 64)
	return s.AddRead(fd)
}

func (s *Select) ModWrite(fd int) error {
	// 删除读
	s.read.Bits[fd/64] &^= 1 << (fd % 64)
	return s.AddWrite(fd)
}

func (s *Select) Remove(fd int) error {
	s.read.Bits[fd/64] &^= 1 << (fd % 64)
	s.write.Bits[fd/64] &^= 1 << (fd % 64)
	return nil
}

// Polling .
func (s *Select) Polling() ([]ExternalEvent, error) {

	read := s.read
	write := s.write
	_, err := syscall.Select(s.max+1, &read, &write, nil, nil)
	if err != nil {
		return nil, err
	}
	external := make([]ExternalEvent, 0)

	for _, fd := range s.fds {

		// 可读
		if s.isFdSet(read, fd) {
			external = append(external, ExternalEvent{
				Fd:     fd,
				Opcode: OpcodeRead,
			})
		}

		// 可写
		if s.isFdSet(write, fd) {
			external = append(external, ExternalEvent{
				Fd:     fd,
				Opcode: OpcodeWrite,
			})
		}
	}

	return external, nil
}

func (s *Select) isFdSet(fdSet syscall.FdSet, fd int) bool {
	return (fdSet.Bits[fd/64] & (1 << (fd % 64))) != 0
}
