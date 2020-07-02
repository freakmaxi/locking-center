package flags

import (
	"fmt"
	"net"

	"github.com/freakmaxi/locking-center/cli/terminal"
)

type execution interface {
	Parse() error
	PrintUsage()

	Name() string
	Execute() error
}

func newExecution(managerAddress string, output terminal.Output, command string, basePath string, args []string, version string) (execution, error) {
	addr, err := net.ResolveTCPAddr("tcp", managerAddress)
	if err != nil {
		return nil, err
	}

	switch command {
	case "keys":
		return NewKeys(addr, output, basePath, args), nil
	case "reset":
		return NewReset(addr, output, basePath, args), nil
	}

	return nil, fmt.Errorf("unsupported command")
}
