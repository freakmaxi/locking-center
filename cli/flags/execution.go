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
		/*case "sh":
		return NewShell(managerAddress, version), nil*/
	}

	return nil, fmt.Errorf("unsupported command")
}

/*
func createTemporary(sources []string, target string) error {
	targetFile, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0777)
	if err != nil {
		return fmt.Errorf("unable to create temporary file: %s", err.Error())
	}
	defer targetFile.Close()

	for _, source := range sources {
		info, err := os.Stat(source)
		if err != nil {
			return fmt.Errorf("unable to read %s", source)
		}
		if info.IsDir() {
			return fmt.Errorf("%s should be file", source)
		}

		if err := combine(source, targetFile); err != nil {
			return err
		}
	}

	return nil
}

func combine(source string, writer io.Writer) error {
	file, err := os.OpenFile(source, os.O_RDONLY, 0666)
	if err != nil {
		return fmt.Errorf("problem on accessing to file: %s", err.Error())
	}
	defer file.Close()

	if _, err := io.Copy(writer, file); err != nil {
		return fmt.Errorf("problem on combining %s to temporary file: %s", source, err.Error())
	}

	return nil
}
*/
