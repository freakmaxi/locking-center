package service

import (
	"fmt"
	"io"
	"net"
	"strings"
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
		return
	}
	if err := m.socketIO.WriteWithTimeout(conn, []byte{'+'}); err != nil {
		fmt.Printf("ERROR: Service failed on success message: address: %s,%s\n", conn.RemoteAddr(), err)
	}
}

func (m *mutex) process(conn net.Conn) error {
	var keySize int8
	if err := m.socketIO.ReadBinaryWithTimeout(conn, &keySize); err != nil {
		return err
	}

	keyBytes := make([]byte, keySize)
	if err := m.socketIO.ReadWithTimeout(conn, keyBytes, len(keyBytes)); err != nil {
		return err
	}
	key := string(keyBytes)

	var action mutexAction
	if err := m.socketIO.ReadBinaryWithTimeout(conn, &action); err != nil {
		return err
	}

	m.socketIO.Idle(conn)

	switch action {
	case maLock: // Lock
		for !m.lock.Lock(key, conn.RemoteAddr()) {
		}
		return nil
	case maUnlock: // Unlock
		m.lock.Unlock(key)
		return nil
	case maResetByKey: // Reset by Key
		m.lock.ResetByKey(key)
		return nil
	case maResetBySource: // Reset by Source Addr
		if len(key) == 0 {
			key = conn.RemoteAddr().String()
			idxColon := strings.Index(key, ":")
			if idxColon > -1 {
				key = key[:idxColon]
			}
		}
		m.lock.ResetBySource(key)
		return nil
	}

	return fmt.Errorf("undefined action")
}
