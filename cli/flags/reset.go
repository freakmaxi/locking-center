package flags

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/freakmaxi/locking-center/cli/errors"
	"github.com/freakmaxi/locking-center/cli/terminal"
)

const resetByKeyRemoteCommand = "RSET"
const resetBySourceRemoteCommand = "RSBS"

type resetCommand struct {
	managerAddress *net.TCPAddr
	output         terminal.Output
	basePath       string
	args           []string

	keys  []string
	byKey bool
}

func NewReset(managerAddress *net.TCPAddr, output terminal.Output, basePath string, args []string) execution {
	return &resetCommand{
		managerAddress: managerAddress,
		output:         output,
		basePath:       basePath,
		args:           args,
		byKey:          true,
	}
}

func (r *resetCommand) Parse() error {
	for len(r.args) > 0 {
		arg := r.args[0]
		switch arg {
		case "-s":
			r.args = r.args[1:]
			r.byKey = false
			continue
		case "-h":
			return errors.ErrShowUsage
		default:
			if strings.Index(arg, "-") == 0 {
				return fmt.Errorf("unsupported argument for reset command")
			}
		}
		break
	}

	if r.byKey && len(r.args) == 0 {
		return fmt.Errorf("reset command needs key/source addr parameter")
	}

	r.keys = make([]string, len(r.args))
	copy(r.keys, r.args)

	return nil
}

func (r *resetCommand) PrintUsage() {
	r.output.Println("  reset       Reset locking key and release all locks.")
	r.output.Println("              Ex: reset locking-key")
	r.output.Println("")
	r.output.Println("arguments:")
	r.output.Println("  -s          reset by source address. use source address as locking-key parameter")
	r.output.Println("  -h          shows this help text")
	r.output.Println("")
	r.output.Refresh()
}

func (r *resetCommand) Name() string {
	return "reset"
}

func (r *resetCommand) Execute() error {
	conn, err := net.DialTCP("tcp", nil, r.managerAddress)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	command := resetByKeyRemoteCommand
	if !r.byKey {
		command = resetBySourceRemoteCommand
	}

	if _, err := conn.Write([]byte(command)); err != nil {
		return err
	}

	if err := binary.Write(conn, binary.LittleEndian, uint32(len(r.keys))); err != nil {
		return err
	}

	if !r.byKey && len(r.keys) == 0 {
		return r.reset(conn, "")
	}

	for _, key := range r.keys {
		if err := r.reset(conn, key); err == nil {
			return err
		}
	}

	return nil
}

func (r *resetCommand) reset(conn *net.TCPConn, key string) error {
	if r.byKey && (len(key) == 0 || len(key) > 128) {
		return fmt.Errorf("key is empty or more than 128 characters: %s", key)
	}

	keySize := uint8(len(key))
	if err := binary.Write(conn, binary.LittleEndian, keySize); err != nil {
		return err
	}

	if _, err := conn.Write([]byte(key)); err != nil {
		return err
	}

	if !r.result(conn) {
		return fmt.Errorf("reseting key is failed: %s", key)
	}

	return nil
}

func (r *resetCommand) result(conn *net.TCPConn) bool {
	res := make([]byte, 1)

	if _, err := io.ReadAtLeast(conn, res, len(res)); err != nil {
		return false
	}

	return string(res) == "+"
}
