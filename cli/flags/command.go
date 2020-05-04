package flags

import (
	"fmt"
	"os"
	"path"

	"github.com/freakmaxi/locking-center/cli/terminal"
)

type Command struct {
	version string

	filename       string
	args           []string
	managerAddress string
	command        execution
}

func NewCommand(version string, args []string) *Command {
	_, filename := path.Split(args[0])

	mrArgs := make([]string, 0)
	if 1 < len(args) {
		mrArgs = args[1:]
	}

	return &Command{
		version:        version,
		filename:       filename,
		args:           mrArgs,
		managerAddress: "localhost:22120",
	}
}

func (c *Command) printUsageHeader() {
	fmt.Printf("Locking-Center (v%s) usage: \n", c.version)
	fmt.Println()
}

func (c *Command) printUsage() {
	c.printUsageHeader()
	fmt.Printf("   %s [options] command [arguments] parameters\n", c.filename)
	fmt.Println()
	fmt.Println("options:")
	fmt.Println("  --manager-address   Points the end point of manager node to work with. Default: localhost:22120")
	fmt.Println("  --help              Prints this usage documentation")
	fmt.Println("  --version           Prints release version")
	fmt.Println()
	fmt.Println("commands:")
	fmt.Println("  keys    List locking keys.")
	fmt.Println("  reset   Reset locking key and release all locks.")
	fmt.Println()
}

func (c *Command) Parse() bool {
	if len(c.args) == 0 {
		c.printUsage()
		return false
	}

	for i := 0; i < len(c.args); i++ {
		arg := c.args[i]

		switch arg {
		case "--manager-address":
			if i+1 == len(c.args) {
				fmt.Println("--manager-address requires value")
				fmt.Println()
				c.printUsage()
				return false
			}

			i++
			c.managerAddress = c.args[i]
			continue
		case "--help":
			c.printUsage()
			return false
		case "--version":
			fmt.Printf("%s\n", c.version)
			return false
		}

		switch arg {
		case "keys", "reset":
			mrArgs := make([]string, 0)
			if i+1 < len(c.args) {
				mrArgs = c.args[i+1:]
			}

			var err error
			c.command, err = newExecution(c.managerAddress, terminal.NewStdOut(), arg, string(os.PathSeparator), mrArgs, c.version)
			if err != nil {
				fmt.Println(err.Error())
				fmt.Println()
				c.printUsage()
				return false
			}

			err = c.command.Parse()
			if err != nil {
				fmt.Println(err.Error())
				fmt.Println()
				c.printUsageHeader()
				c.command.PrintUsage()
				return false
			}

			return true
		}
	}

	c.printUsage()
	return false
}

func (c *Command) Execute() error {
	return c.command.Execute()
}
