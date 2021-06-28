package service

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/freakmaxi/locking-center/mutex/common"
)

const commandBuffer = 4             // 4b
const defaultTransferSpeed = 625000 // bytes/s

type Manager interface {
	Listen(wg *sync.WaitGroup) error
}

type manager struct {
	address  *net.TCPAddr
	lock     *common.Lock
	socketIO *SocketIO

	listener *net.TCPListener
	quiting  bool
}

func NewManager(address string, lock *common.Lock) (Manager, error) {
	if len(address) == 0 {
		return nil, fmt.Errorf("address should be defined")
	}
	addr, _ := net.ResolveTCPAddr("tcp", address)

	return &manager{
		address:  addr,
		lock:     lock,
		socketIO: NewSocketIO(),
	}, nil
}

func (m *manager) Listen(wg *sync.WaitGroup) error {
	var err error
	m.listener, err = net.ListenTCP("tcp", m.address)
	if err != nil {
		return err
	}

	fmt.Printf("INFO: Manager Service has started listening at %s\n", m.address.String())

	go func() {
		defer wg.Done()

		for !m.quiting {
			c, err := m.listener.Accept()
			if err != nil {
				fmt.Printf("ERROR: Unable to accept connection: %s\n", err.Error())
				continue
			}
			go m.handler(c)
		}
	}()

	return nil
}

func (m *manager) handler(conn net.Conn) {
	defer func() { _ = conn.Close() }()

	buffer := make([]byte, commandBuffer)

	if err := m.socketIO.ReadWithTimeout(conn, buffer, len(buffer)); err != nil {
		fmt.Printf("ERROR: Stream unable to read: Connection: %s, %s\n", conn.RemoteAddr().String(), err.Error())
		return
	}

	if err := m.process(string(buffer), conn); err != nil {
		if err != io.EOF {
			fmt.Printf("ERROR: Service process is failed: address: %s,%s\n", conn.RemoteAddr(), err)
		}
		if err := m.socketIO.WriteWithTimeout(conn, []byte{'-'}); err != nil {
			fmt.Printf("ERROR: Service failed on unsuccess message: address: %s,%s\n", conn.RemoteAddr(), err)
		}
		return
	}
	if err := m.socketIO.WriteWithTimeout(conn, []byte{'+'}); err != nil {
		fmt.Printf("ERROR: Service failed on success message: address: %s,%s\n", conn.RemoteAddr(), err)
	}
}

func (m *manager) process(command string, conn net.Conn) error {
	switch command {
	case "KEYS":
		return m.keys(conn)
	case "RSET":
		return m.reset(conn, true)
	case "RSBS":
		return m.reset(conn, false)
	default:
		return fmt.Errorf("not a meaningful command")
	}
}

func (m *manager) keys(conn net.Conn) error {
	reports := m.lock.Keys()

	if err := m.socketIO.WriteBinaryWithTimeout(conn, uint32(len(reports))); err != nil {
		return err
	}

	for _, report := range reports {
		keySize := uint8(len(report.Key))
		if err := m.socketIO.WriteBinaryWithTimeout(conn, keySize); err != nil {
			return err
		}

		keyBytes := []byte(report.Key)
		if err := m.socketIO.WriteWithTimeout(conn, keyBytes); err != nil {
			return err
		}

		sourceAddrSize := uint8(len(report.Current.SourceAddr))
		if err := m.socketIO.WriteBinaryWithTimeout(conn, sourceAddrSize); err != nil {
			return err
		}

		sourceAddrBytes := []byte(report.Current.SourceAddr)
		if err := m.socketIO.WriteWithTimeout(conn, sourceAddrBytes); err != nil {
			return err
		}

		endPointSize := uint8(len(report.Current.RemoteAddr.String()))
		if err := m.socketIO.WriteBinaryWithTimeout(conn, endPointSize); err != nil {
			return err
		}

		endPointBytes := []byte(report.Current.RemoteAddr.String())
		if err := m.socketIO.WriteWithTimeout(conn, endPointBytes); err != nil {
			return err
		}

		if err := m.socketIO.WriteBinaryWithTimeout(conn, report.Current.Stamp.Unix()); err != nil {
			return err
		}
	}

	return nil
}

func (m *manager) reset(conn net.Conn, byKey bool) error {
	var resetKeysCount uint32
	if err := m.socketIO.ReadBinaryWithTimeout(conn, &resetKeysCount); err != nil {
		return err
	}

	if !byKey && resetKeysCount == 0 {
		key := common.ExtractSourceAddr(conn)

		m.lock.ResetBySource(key)

		return m.socketIO.WriteWithTimeout(conn, []byte("+"))
	}

	for ; resetKeysCount > 0; resetKeysCount-- {
		var keySize uint8
		if err := m.socketIO.ReadBinaryWithTimeout(conn, &keySize); err != nil {
			return err
		}

		keyBytes := make([]byte, keySize)
		if err := m.socketIO.ReadWithTimeout(conn, keyBytes, len(keyBytes)); err != nil {
			return err
		}

		key := string(keyBytes)

		m.socketIO.Idle(conn)

		if byKey {
			m.lock.ResetByKey(key)
		} else {
			m.lock.ResetBySource(key)
		}

		if err := m.socketIO.WriteWithTimeout(conn, []byte("+")); err != nil {
			return err
		}
	}

	return nil
}
