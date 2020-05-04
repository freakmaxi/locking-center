package flags

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"github.com/freakmaxi/locking-center/cli/terminal"
)

const keysRemoteCommand = "KEYS"

type keysCommand struct {
	managerAddress *net.TCPAddr
	output         terminal.Output
	basePath       string
	args           []string

	//listing bool
	//source  string
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
	/*for len(l.args) > 0 {
		arg := l.args[0]
		switch arg {
		case "-l":
			l.args = l.args[1:]
			l.listing = true
			continue
		case "-u":
			l.args = l.args[1:]
			l.usage = true
			continue
		case "-h":
			return errors.ErrShowUsage
		default:
			if strings.Index(arg, "-") == 0 {
				return fmt.Errorf("unsupported argument for ls command")
			}
		}
		break
	}

	if len(l.args) > 1 {
		return fmt.Errorf("ls command needs optionally source parameter")
	}

	l.source = l.basePath
	if len(l.args) == 1 {
		if !filepath.IsAbs(l.args[0]) {
			l.source = path.Join(l.basePath, l.args[0])
		} else {
			l.source = l.args[0]
		}
	}
	*/
	return nil
}

func (k *keysCommand) PrintUsage() {
	k.output.Println("  keys        List locking keys.")
	k.output.Println("")
	/*l.output.Println("arguments:")
	l.output.Println("  -l          shows in a listing format")
	l.output.Println("  -u          calculate the size of folders")
	l.output.Println("")
	l.output.Println("marking:")
	l.output.Println("  d           folder")
	l.output.Println("  -           file")
	l.output.Println("  •           locked")
	l.output.Println("  ↯           zombie")
	l.output.Println("")*/
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
	defer conn.Close()

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

		fmt.Println(string(keyBytes))
	}

	return nil
	/*if strings.Index(l.source, local) == 0 {
		return fmt.Errorf("please use O/S native commands to list files/folders")
	}

	anim := common.NewAnimation(l.output, "processing...")
	anim.Start()

	folder, err := dfs.List(l.headAddresses, l.source, l.usage)
	if err != nil {
		anim.Cancel()
		return err
	}
	anim.Stop()

	if l.listing {
		l.printAsList(folder)
	} else {
		l.printAsSummary(folder)
	}
	return nil*/
}

/*func (l *listCommand) printAsSummary(folder *common.Folder) {
	for _, f := range folder.Folders {
		if l.usage {
			l.output.Printf("> %s (%s)   ", f.Name, l.sizeToString(f.Size))
			continue
		}
		l.output.Printf("> %s   ", f.Name)
	}
	for _, f := range folder.Files {
		l.output.Printf("%s   ", f.Name)
	}
	l.output.Println("")
	l.output.Refresh()
}

func (l *listCommand) printAsList(folder *common.Folder) {
	total := len(folder.Folders) + len(folder.Files)

	if l.usage && total > 1 {
		l.output.Printf("total %d (%s)\n", total, l.sizeToString(folder.Size))
	} else {
		l.output.Printf("total %d\n", total)
	}

	for _, f := range folder.Folders {
		l.output.Printf("d %7v %s %s\n", l.sizeToString(f.Size), f.Created.Format("2006 Jan 02 03:04"), f.Name)
	}

	for _, f := range folder.Files {
		name := f.Name
		lockChar := "-"
		if f.Zombie {
			lockChar = "↯"
		} else if f.Locked() {
			lockChar = "•"
			name = fmt.Sprintf("%s (locked till %s)", name, f.Lock.Till.Local().Format("2006 Jan 02 03:04"))
		}
		l.output.Printf("%s %7v %s %s\n", lockChar, l.sizeToString(f.Size), f.Modified.Local().Format("2006 Jan 02 03:04"), name)
	}

	l.output.Refresh()
}

func (l *listCommand) sizeToString(size uint64) string {
	calculatedSize := size
	divideCount := 0
	for {
		calculatedSizeString := strconv.FormatUint(calculatedSize, 10)
		if len(calculatedSizeString) < 6 {
			break
		}
		calculatedSize /= 1024
		divideCount++
	}

	switch divideCount {
	case 0:
		return fmt.Sprintf("%sb", strconv.FormatUint(calculatedSize, 10))
	case 1:
		return fmt.Sprintf("%skb", strconv.FormatUint(calculatedSize, 10))
	case 2:
		return fmt.Sprintf("%smb", strconv.FormatUint(calculatedSize, 10))
	case 3:
		return fmt.Sprintf("%sgb", strconv.FormatUint(calculatedSize, 10))
	case 4:
		return fmt.Sprintf("%stb", strconv.FormatUint(calculatedSize, 10))
	}

	return "N/A"
}
*/
