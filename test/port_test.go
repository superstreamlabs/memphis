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
	"net"
	"strconv"
	"testing"

	"memphis-broker/server"
)

func TestResolveRandomPort(t *testing.T) {
	opts := &server.Options{Host: "127.0.0.1", Port: server.RANDOM_PORT, NoSigs: true}
	s := RunServer(opts)
	defer s.Shutdown()

	addr := s.Addr()
	_, port, err := net.SplitHostPort(addr.String())
	if err != nil {
		t.Fatalf("Expected no error: Got %v\n", err)
	}

	portNum, err := strconv.Atoi(port)
	if err != nil {
		t.Fatalf("Expected no error: Got %v\n", err)
	}

	if portNum == server.DEFAULT_PORT {
		t.Fatalf("Expected server to choose a random port\nGot: %d", server.DEFAULT_PORT)
	}

	if portNum == server.RANDOM_PORT {
		t.Fatalf("Expected server to choose a random port\nGot: %d", server.RANDOM_PORT)
	}

	if opts.Port != portNum {
		t.Fatalf("Options port (%d) should have been overridden by chosen random port (%d)",
			opts.Port, portNum)
	}
}
