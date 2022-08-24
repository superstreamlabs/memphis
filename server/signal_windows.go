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
package server

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

// Signal Handling
func (s *Server) handleSignals() {
	if s.getOpts().NoSigs {
		return
	}
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		for sig := range c {
			s.Debugf("Trapped %q signal", sig)
			s.Shutdown()
			os.Exit(0)
		}
	}()
}

// ProcessSignal sends the given signal command to the running nats-server service.
// If service is empty, this signals the "nats-server" service. This returns an
// error is the given service is not running or the command is invalid.
func ProcessSignal(command Command, service string) error {
	if service == "" {
		service = serviceName
	}

	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := m.OpenService(service)
	if err != nil {
		return fmt.Errorf("could not access service: %v", err)
	}
	defer s.Close()

	var (
		cmd svc.Cmd
		to  svc.State
	)

	switch command {
	case CommandStop, CommandQuit:
		cmd = svc.Stop
		to = svc.Stopped
	case CommandReopen:
		cmd = reopenLogCmd
		to = svc.Running
	case CommandReload:
		cmd = svc.ParamChange
		to = svc.Running
	case commandLDMode:
		cmd = ldmCmd
		to = svc.Running
	default:
		return fmt.Errorf("unknown signal %q", command)
	}

	status, err := s.Control(cmd)
	if err != nil {
		return fmt.Errorf("could not send control=%d: %v", cmd, err)
	}

	timeout := time.Now().Add(10 * time.Second)
	for status.State != to {
		if timeout.Before(time.Now()) {
			return fmt.Errorf("timeout waiting for service to go to state=%d", to)
		}
		time.Sleep(300 * time.Millisecond)
		status, err = s.Query()
		if err != nil {
			return fmt.Errorf("could not retrieve service status: %v", err)
		}
	}

	return nil
}
