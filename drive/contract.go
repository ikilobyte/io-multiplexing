package drive

type Poller interface {
	AddRead(fd int) error
	AddWrite(fd int) error
	ModRead(fd int) error
	ModWrite(fd int) error
	Remove(fd int) error
	Polling() ([]ExternalEvent, error)
}
