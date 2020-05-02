package main

import (
	"fmt"
	"os"
	"regexp"

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

	master, err := service.NewMaster(bindAddr, common.NewLock())
	if err != nil {
		fmt.Printf("ERROR: Service unable to be prepared: %s", err)
		os.Exit(5)
	}

	if err := master.Listen(); err != nil {
		fmt.Printf("ERROR: Service unable to be started: %s", err)
		os.Exit(10)
	}
}
