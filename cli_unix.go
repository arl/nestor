//go:build unix
// +build unix

package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
)

func installStacktraceHandler() {
	progname := filepath.Base(os.Args[0])
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGUSR2)

	go func() {
		for {
			<-sig

			buf := make([]byte, 1<<20)
			stackSize := runtime.Stack(buf, true)
			pattern := fmt.Sprintf("stacktrace.%s.*", progname)

			f, err := os.CreateTemp("", pattern)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to create stack trace file: %s\n", err)
			}
			f.Write(buf[:stackSize])
			f.Close()
			fmt.Fprintf(os.Stderr, "stack trace created: %s\n", f.Name())
		}
	}()
}
