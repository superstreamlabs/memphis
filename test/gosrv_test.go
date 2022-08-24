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
package test

import (
	"fmt"
	"net"
	"runtime"
	"testing"
	"time"
)

func TestSimpleGoServerShutdown(t *testing.T) {
	base := runtime.NumGoroutine()
	opts := DefaultTestOptions
	opts.Port = -1
	s := RunServer(&opts)
	s.Shutdown()
	checkFor(t, time.Second, 100*time.Millisecond, func() error {
		delta := (runtime.NumGoroutine() - base)
		if delta > 1 {
			return fmt.Errorf("%d go routines still exist post Shutdown()", delta)
		}
		return nil
	})
}

func TestGoServerShutdownWithClients(t *testing.T) {
	base := runtime.NumGoroutine()
	opts := DefaultTestOptions
	opts.Port = -1
	s := RunServer(&opts)
	addr := s.Addr().(*net.TCPAddr)
	for i := 0; i < 50; i++ {
		createClientConn(t, "127.0.0.1", addr.Port)
	}
	s.Shutdown()
	// Wait longer for client connections
	time.Sleep(1 * time.Second)
	delta := (runtime.NumGoroutine() - base)
	// There may be some finalizers or IO, but in general more than
	// 2 as a delta represents a problem.
	if delta > 2 {
		t.Fatalf("%d Go routines still exist post Shutdown()", delta)
	}
}

func TestGoServerMultiShutdown(t *testing.T) {
	opts := DefaultTestOptions
	opts.Port = -1
	s := RunServer(&opts)
	s.Shutdown()
	s.Shutdown()
}
