package terminal

/*
typedef struct termios Termios;
#include <termios.h>
#include <string.h>
#include <errno.h>

int getErrno() {
  return errno;
}
*/
import "C"
import (
	"errors"
)

func enableFlag(flag uint) error {
	var attr C.Termios
	if C.tcgetattr(0, &attr) == -1 {
		return errors.New(C.GoString(C.strerror(C.getErrno())))
	}
	attr.c_lflag |= C.tcflag_t(flag)
	if C.tcsetattr(0, 0, &attr) == -1 {
		return errors.New(C.GoString(C.strerror(C.getErrno())))
	}
	return nil
}

func disableFlag(flag uint) error {
	var attr C.Termios
	if C.tcgetattr(0, &attr) == -1 {
		return errors.New(C.GoString(C.strerror(C.getErrno())))
	}
	attr.c_lflag &= ^C.tcflag_t(flag)
	if C.tcsetattr(0, 0, &attr) == -1 {
		return errors.New(C.GoString(C.strerror(C.getErrno())))
	}
	return nil
}

func EnableEchoMode() error {
	return enableFlag(C.ECHO)
}

func DisableEchoMode() error {
	return disableFlag(C.ECHO)
}

func EnableCannonicalMode() error {
	return enableFlag(C.ICANON)
}

func DisableCannonicalMode() error {
	return disableFlag(C.ICANON)
}
