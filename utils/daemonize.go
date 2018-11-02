// +build !windows

package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/VividCortex/godaemon"
	"github.com/nightlyone/lockfile"
)

const CanDaemonize = true

func ResolvePath(path string) string {

	if filepath.IsAbs(path) {
		return path
	}

	return filepath.Join(os.Getenv("__DAEMON_CWD"), path)
}

func Daemonize(logFilePath, pidFilePath string) {

	if os.Getenv("__DAEMON_CWD") == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Cannot determine working directory:", err)
			os.Exit(1)
		}
		os.Setenv("__DAEMON_CWD", cwd)
	}

	logFile, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not open local log file:", err)
		os.Exit(1)
	}

	stdout, stderr, err := godaemon.MakeDaemon(&godaemon.DaemonAttr{CaptureOutput: true})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not Daemonize:", err)
		os.Exit(1)
	}

	go func() {
		io.Copy(logFile, stdout)
	}()
	go func() {
		io.Copy(logFile, stderr)
	}()

	lock, err := lockfile.New(pidFilePath)
	if err != nil {
		fmt.Printf("Error locking: %v\n", err)
		os.Exit(1)
	}
	err = lock.TryLock()
	if err != nil {
		fmt.Printf("Cannot lock \"%v\": %v\n", lock, err)
		os.Exit(1)
	}

}
