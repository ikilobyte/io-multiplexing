package drive

type Opcode string

const (
	OpcodeRead  Opcode = "read"
	OpcodeWrite Opcode = "write"
)

type ExternalEvent struct {
	Fd     int
	Opcode Opcode
}
