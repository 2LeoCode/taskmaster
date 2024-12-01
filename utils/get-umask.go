package utils

import "syscall"

func GetUmask() int {
	umask := syscall.Umask(0)
	syscall.Umask(umask)
	return umask
}
