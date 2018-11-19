// +build !windows

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

// Mostly copied from Docker
func dumpStacks() {
	var (
		buf       []byte
		stackSize int
	)

	// Continually grab the trace until the buffer size exceeds its length
	bufferLen := 16384
	for stackSize == len(buf) {
		buf = make([]byte, bufferLen)
		stackSize = runtime.Stack(buf, true)
		bufferLen *= 2
	}
	buf = buf[:stackSize]

	f, err := ioutil.TempFile("", "r_s_stacktrace")
	defer f.Close()
	if err == nil {
		f.WriteString(string(buf))
	}

}

func AddSignalHandlers(c *Config) {
	signals := make(chan os.Signal, 100)

	signal.Notify(signals,
		syscall.SIGUSR1,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		for sig := range signals {
			go func(sig os.Signal) {
				fmt.Fprintf(os.Stderr, "Handling signal: %v\n", sig)
				switch sig {
				case syscall.SIGUSR1:
					dumpStacks()
				case syscall.SIGHUP:
					dumpConfig(c)
				case syscall.SIGINT, syscall.SIGTERM:
					// removePidFile()
					os.Exit(128 + int(sig.(syscall.Signal)))
				case syscall.SIGQUIT:
					//removePidFile
					os.Exit(0)
				}
			}(sig)
		}
	}()
}
