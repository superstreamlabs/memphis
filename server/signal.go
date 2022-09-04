// Credit for The NATS.IO Authors
// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

//go:build !windows && !wasm
// +build !windows,!wasm

package server

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

var processName = "nats-server"

// SetProcessName allows to change the expected name of the process.
func SetProcessName(name string) {
	processName = name
}

// Signal Handling
func (s *Server) handleSignals() {
	if s.getOpts().NoSigs {
		return
	}
	c := make(chan os.Signal, 1)

	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGHUP)

	go func() {
		for {
			select {
			case sig := <-c:
				s.Debugf("Trapped %q signal", sig)
				switch sig {
				case syscall.SIGINT:
					s.Shutdown()
					os.Exit(0)
				case syscall.SIGTERM:
					// Shutdown unless graceful shutdown already in progress.
					s.mu.Lock()
					ldm := s.ldm
					s.mu.Unlock()

					if !ldm {
						s.Shutdown()
						os.Exit(1)
					}
				case syscall.SIGUSR1:
					// File log re-open for rotating file logs.
					s.ReOpenLogFile()
				case syscall.SIGUSR2:
					go s.lameDuckMode()
				case syscall.SIGHUP:
					// Config reload.
					if err := s.Reload(); err != nil {
						s.Errorf("Failed to reload server configuration: %s", err)
					}
				}
			case <-s.quitCh:
				return
			}
		}
	}()
}

// ProcessSignal sends the given signal command to the given process. If pidStr
// is empty, this will send the signal to the single running instance of
// nats-server. If multiple instances are running, it returns an error. This returns
// an error if the given process is not running or the command is invalid.
func ProcessSignal(command Command, pidStr string) error {
	var pid int
	if pidStr == "" {
		pids, err := resolvePids()
		if err != nil {
			return err
		}
		if len(pids) == 0 {
			return fmt.Errorf("no %s processes running", processName)
		}
		if len(pids) > 1 {
			errStr := fmt.Sprintf("multiple %s processes running:\n", processName)
			prefix := ""
			for _, p := range pids {
				errStr += fmt.Sprintf("%s%d", prefix, p)
				prefix = "\n"
			}
			return errors.New(errStr)
		}
		pid = pids[0]
	} else {
		p, err := strconv.Atoi(pidStr)
		if err != nil {
			return fmt.Errorf("invalid pid: %s", pidStr)
		}
		pid = p
	}

	var err error
	switch command {
	case CommandStop:
		err = kill(pid, syscall.SIGKILL)
	case CommandQuit:
		err = kill(pid, syscall.SIGINT)
	case CommandReopen:
		err = kill(pid, syscall.SIGUSR1)
	case CommandReload:
		err = kill(pid, syscall.SIGHUP)
	case commandLDMode:
		err = kill(pid, syscall.SIGUSR2)
	case commandTerm:
		err = kill(pid, syscall.SIGTERM)
	default:
		err = fmt.Errorf("unknown signal %q", command)
	}
	return err
}

// resolvePids returns the pids for all running nats-server processes.
func resolvePids() ([]int, error) {
	// If pgrep isn't available, this will just bail out and the user will be
	// required to specify a pid.
	output, err := pgrep()
	if err != nil {
		switch err.(type) {
		case *exec.ExitError:
			// ExitError indicates non-zero exit code, meaning no processes
			// found.
			break
		default:
			return nil, errors.New("unable to resolve pid, try providing one")
		}
	}
	var (
		myPid   = os.Getpid()
		pidStrs = strings.Split(string(output), "\n")
		pids    = make([]int, 0, len(pidStrs))
	)
	for _, pidStr := range pidStrs {
		if pidStr == "" {
			continue
		}
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			return nil, errors.New("unable to resolve pid, try providing one")
		}
		// Ignore the current process.
		if pid == myPid {
			continue
		}
		pids = append(pids, pid)
	}
	return pids, nil
}

var kill = func(pid int, signal syscall.Signal) error {
	return syscall.Kill(pid, signal)
}

var pgrep = func() ([]byte, error) {
	return exec.Command("pgrep", processName).Output()
}
