package shell

import (
	"fmt"
	"sync"
)

var commandLock sync.Mutex

var command string
var cursor int

func DisplayCommand() {
	commandLock.Lock()
	fmt.Printf("\033[2K\r> %s", command)
	if len(command) != cursor {
		fmt.Printf("\033[%dD", len(command)-cursor)
	}
	commandLock.Unlock()
}
