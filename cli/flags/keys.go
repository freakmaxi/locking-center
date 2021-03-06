package flags

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/freakmaxi/locking-center/cli/errors"
	"github.com/freakmaxi/locking-center/cli/terminal"
)

const keysRemoteCommand = "KEYS"

type keysCommand struct {
	managerAddress *net.TCPAddr
	output         terminal.Output
	basePath       string
	args           []string

	detailed bool
}

func NewKeys(managerAddress *net.TCPAddr, output terminal.Output, basePath string, args []string) execution {
	return &keysCommand{
		managerAddress: managerAddress,
		output:         output,
		basePath:       basePath,
		args:           args,
	}
}

func (k *keysCommand) Parse() error {
	for len(k.args) > 0 {
		arg := k.args[0]
		switch arg {
		case "-d":
			k.args = k.args[1:]
			k.detailed = true
			continue
		case "-h":
			return errors.ErrShowUsage
		default:
			if strings.Index(arg, "-") == 0 {
				return fmt.Errorf("unsupported argument for keys command")
			}
		}
		break
	}

	if len(k.args) > 0 {
		return fmt.Errorf("keys command takes only optional modifier arguments")
	}

	return nil
}

func (k *keysCommand) PrintUsage() {
	k.output.Println("  keys        List locking keys.")
	k.output.Println("")
	k.output.Println("arguments:")
	k.output.Println("  -d          show detailed usage of keys")
	k.output.Println("")
	k.output.Refresh()
}

func (k *keysCommand) Name() string {
	return "keys"
}

func (k *keysCommand) Execute() error {
	conn, err := net.DialTCP("tcp", nil, k.managerAddress)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	if _, err := conn.Write([]byte(keysRemoteCommand)); err != nil {
		return err
	}

	var keysCount int32
	if err := binary.Read(conn, binary.LittleEndian, &keysCount); err != nil {
		return err
	}

	for ; keysCount > 0; keysCount-- {
		var keySize uint8
		if err := binary.Read(conn, binary.LittleEndian, &keySize); err != nil {
			return err
		}

		keyBytes := make([]byte, keySize)
		if _, err := io.ReadAtLeast(conn, keyBytes, len(keyBytes)); err != nil {
			return err
		}

		var sourceAddrSize uint8
		if err := binary.Read(conn, binary.LittleEndian, &sourceAddrSize); err != nil {
			return err
		}

		sourceAddrBytes := make([]byte, sourceAddrSize)
		if _, err := io.ReadAtLeast(conn, sourceAddrBytes, len(sourceAddrBytes)); err != nil {
			return err
		}

		var endPointSize uint8
		if err := binary.Read(conn, binary.LittleEndian, &endPointSize); err != nil {
			return err
		}

		endPointBytes := make([]byte, endPointSize)
		if _, err := io.ReadAtLeast(conn, endPointBytes, len(endPointBytes)); err != nil {
			return err
		}

		var unixTime int64
		if err := binary.Read(conn, binary.LittleEndian, &unixTime); err != nil {
			return err
		}

		if k.detailed {
			t := time.Unix(unixTime, 0)
			r := strings.Split(string(endPointBytes), ":")
			d := time.Now().Sub(t)

			fmt.Printf(
				"%15s:%-5s -> %s (%9.3fs) %s (%s)\n",
				r[0],
				r[1],
				t.Local().Format("2006 Jan 02 15:04:03"),
				d.Seconds(),
				string(keyBytes),
				string(sourceAddrBytes),
			)

			continue
		}

		fmt.Println(string(keyBytes))
	}

	return nil
}
