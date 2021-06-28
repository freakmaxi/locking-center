package service

import (
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/freakmaxi/locking-center/mutex/common"
)

type mutexAction byte

var (
	maLock          mutexAction = 1
	maUnlock        mutexAction = 2
	maResetByKey    mutexAction = 3
	maResetBySource mutexAction = 4
)

type Mutex interface {
	Listen(wg *sync.WaitGroup) error
}

type mutex struct {
	address  *net.TCPAddr
	lock     *common.Lock
	socketIO *SocketIO

	listener *net.TCPListener
	quiting  bool
}

func NewMutex(address string, lock *common.Lock) (Mutex, error) {
	if len(address) == 0 {
		return nil, fmt.Errorf("address should be defined")
	}
	addr, _ := net.ResolveTCPAddr("tcp4", address)

	return &mutex{
		address:  addr,
		lock:     lock,
		socketIO: NewSocketIO(),
	}, nil
}

func (m *mutex) Listen(wg *sync.WaitGroup) error {
	var err error
	m.listener, err = net.ListenTCP("tcp", m.address)
	if err != nil {
		return err
	}

	fmt.Printf("INFO: Mutex Service has started listening at %s\n", m.address.String())

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

func (m *mutex) handler(conn net.Conn) {
	defer func() { _ = conn.Close() }()

	if err := m.process(conn); err != nil {
		if err != io.EOF {
			fmt.Printf("ERROR: Service process is failed: address: %s,%s\n", conn.RemoteAddr(), err)
		}
		if err := m.socketIO.WriteWithTimeout(conn, []byte{'-'}); err != nil {
			fmt.Printf("ERROR: Service failed on unsuccess message: address: %s,%s\n", conn.RemoteAddr(), err)
		}
	}
}

func (m *mutex) success(conn net.Conn) bool {
	if err := m.socketIO.WriteWithTimeout(conn, []byte{'+'}); err != nil {
		fmt.Printf("ERROR: Service failed on success message: address: %s,%s\n", conn.RemoteAddr(), err)
		return false
	}
	return true
}

func (m *mutex) process(conn net.Conn) error {
	var action mutexAction
	if err := m.socketIO.ReadBinaryWithTimeout(conn, &action); err != nil {
		return err
	}

	switch action {
	case maLock:
		return m.cmdLock(conn)
	case maUnlock:
		return m.cmdUnlock(conn)
	case maResetByKey:
		return m.cmdResetByKey(conn)
	case maResetBySource:
		return m.cmdResetBySource(conn)
	default:
		return fmt.Errorf("undefined action")
	}
}

func (m *mutex) readString(conn net.Conn) (*string, error) {
	var valueSize int8
	if err := m.socketIO.ReadBinaryWithTimeout(conn, &valueSize); err != nil {
		return nil, err
	}

	valueBytes := make([]byte, valueSize)
	if err := m.socketIO.ReadWithTimeout(conn, valueBytes, len(valueBytes)); err != nil {
		return nil, err
	}

	value := string(valueBytes)

	return &value, nil
}

func (m *mutex) cmdLock(conn net.Conn) error {
	key, err := m.readString(conn)
	if err != nil {
		return err
	}

	sourceAddrPtr, err := m.readString(conn)
	if err != nil {
		return err
	}
	sourceAddr := *sourceAddrPtr

	if len(sourceAddr) == 0 {
		sourceAddr = common.ExtractSourceAddr(conn)
	}

	m.socketIO.Idle(conn)

	for !m.lock.Lock(*key, sourceAddr, conn.RemoteAddr()) {
	}

	// If connection is closed before the answer, cancel the lock
	if !m.success(conn) {
		m.lock.Unlock(*key)
	}

	return nil
}

func (m *mutex) cmdUnlock(conn net.Conn) error {
	key, err := m.readString(conn)
	if err != nil {
		return err
	}

	m.socketIO.Idle(conn)

	m.lock.Unlock(*key)
	m.success(conn)

	return nil
}

func (m *mutex) cmdResetByKey(conn net.Conn) error {
	key, err := m.readString(conn)
	if err != nil {
		return err
	}

	m.socketIO.Idle(conn)
	m.lock.ResetByKey(*key)
	m.success(conn)

	return nil
}

func (m *mutex) cmdResetBySource(conn net.Conn) error {
	sourceAddrPtr, err := m.readString(conn)
	if err != nil {
		return err
	}
	sourceAddr := *sourceAddrPtr

	m.socketIO.Idle(conn)

	if len(sourceAddr) == 0 {
		sourceAddr = common.ExtractSourceAddr(conn)
	}

	m.lock.ResetBySource(sourceAddr)
	m.success(conn)

	return nil
}
