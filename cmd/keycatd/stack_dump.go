// +build !windows

package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

func enableStackDump() {
	dump := make(chan os.Signal)
	go func() {
		stack := make([]byte, 16*1024)
		for _ = range dump {
			n := runtime.Stack(stack, true)
			fmt.Fprintf(os.Stderr, "==== %s\n%s\n====\n", time.Now(), stack[0:n])
		}
	}()
	signal.Notify(dump, syscall.SIGUSR1)
}

func init() {
	enableStackDump()
}
