// +build linux darwin

package main

import "syscall"

func init() {
	syscall.Umask(077)
}
