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
	"bufio"
	"fmt"
	"net"
	"testing"
	"time"
)

const PING_CLIENT_PORT = 11228

var DefaultPingOptions = Options{
	Host:         "127.0.0.1",
	Port:         PING_CLIENT_PORT,
	NoLog:        true,
	NoSigs:       true,
	PingInterval: 50 * time.Millisecond,
}

func TestPing(t *testing.T) {
	o := DefaultPingOptions
	o.DisableShortFirstPing = true
	s := RunServer(&o)
	defer s.Shutdown()

	c, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", PING_CLIENT_PORT))
	if err != nil {
		t.Fatalf("Error connecting: %v", err)
	}
	defer c.Close()
	br := bufio.NewReader(c)
	// Wait for INFO
	br.ReadLine()
	// Send CONNECT
	c.Write([]byte("CONNECT {\"verbose\":false}\r\nPING\r\n"))
	// Wait for first PONG
	br.ReadLine()
	// Wait for PING
	start := time.Now()
	for i := 0; i < 3; i++ {
		l, _, err := br.ReadLine()
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if string(l) != "PING" {
			t.Fatalf("Expected PING, got %q", l)
		}
		if dur := time.Since(start); dur < 25*time.Millisecond || dur > 75*time.Millisecond {
			t.Fatalf("Pings duration off: %v", dur)
		}
		c.Write([]byte(pongProto))
		start = time.Now()
	}
}
