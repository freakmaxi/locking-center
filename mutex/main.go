package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/freakmaxi/locking-center/mutex/common"
	"github.com/freakmaxi/locking-center/mutex/service"
)

func main() {
	fmt.Println("INFO: ------------ Starting Locking Center ------------")

	bindAddr := os.Getenv("BIND_ADDRESS")
	if matched, err := regexp.MatchString(`:\d{1,5}$`, bindAddr); err != nil || !matched {
		bindAddr = fmt.Sprintf("%s:22119", bindAddr)
	}
	fmt.Printf("INFO: BIND_ADDRESS: %s\n", bindAddr)

	managerBindAddr := bindAddr
	managerBindAddrParts :=
		strings.Split(managerBindAddr, ":")
	if len(managerBindAddrParts) != 2 {
		fmt.Println("ERROR: Unable to calculate manager bind address")
		os.Exit(2)
	}

	managerPort, err := strconv.ParseUint(managerBindAddrParts[1], 10, 64)
	if err != nil {
		fmt.Printf("ERROR: BIND_ADDRESS is in wrong format: %s\n", err)
		os.Exit(3)
	}

	if managerPort >= uint64(^uint16(0)) {
		fmt.Println("ERROR: BIND_ADDRESS port is at the edge")
		os.Exit(3)
	}
	managerBindAddr = fmt.Sprintf("%s:%d", managerBindAddrParts[0], managerPort+1)

	wg := &sync.WaitGroup{}
	lock := common.NewLock()

	master, err := service.NewMaster(bindAddr, lock)
	if err != nil {
		fmt.Printf("ERROR: Service unable to be prepared: %s\n", err)
		os.Exit(5)
	}

	wg.Add(1)
	if err := master.Listen(wg); err != nil {
		fmt.Printf("ERROR: Service unable to be started: %s\n", err)
		os.Exit(10)
	}

	manager, err := service.NewManager(managerBindAddr, lock)
	if err != nil {
		fmt.Printf("ERROR: Service unable to be prepared: %s\n", err)
		os.Exit(15)
	}

	wg.Add(1)
	if err := manager.Listen(wg); err != nil {
		fmt.Printf("ERROR: Service unable to be started: %s\n", err)
		os.Exit(20)
	}

	wg.Wait()
}
