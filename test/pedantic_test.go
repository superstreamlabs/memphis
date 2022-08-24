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
	"testing"

	"memphis-broker/server"
)

func runPedanticServer() *server.Server {
	opts := DefaultTestOptions

	opts.NoLog = false
	opts.Trace = true

	opts.Port = PROTO_TEST_PORT
	return RunServer(&opts)
}

func TestPedanticSub(t *testing.T) {
	s := runPedanticServer()
	defer s.Shutdown()

	c := createClientConn(t, "127.0.0.1", PROTO_TEST_PORT)
	defer c.Close()

	send := sendCommand(t, c)
	expect := expectCommand(t, c)
	doConnect(t, c, false, true, false)

	// Ping should still be same
	send("PING\r\n")
	expect(pongRe)

	// Test malformed subjects for SUB
	// Sub can contain wildcards, but
	// subject must still be legit.

	// Empty terminal token
	send("SUB foo. 1\r\n")
	expect(errRe)

	// Empty beginning token
	send("SUB .foo. 1\r\n")
	expect(errRe)

	// Empty middle token
	send("SUB foo..bar 1\r\n")
	expect(errRe)

	// Bad non-terminal FWC
	send("SUB foo.>.bar 1\r\n")
	buf := expect(errRe)

	// Check that itr is 'Invalid Subject'
	matches := errRe.FindAllSubmatch(buf, -1)
	if len(matches) != 1 {
		t.Fatal("Wanted one overall match")
	}
	if string(matches[0][1]) != "'Invalid Subject'" {
		t.Fatalf("Expected 'Invalid Subject', got %s", string(matches[0][1]))
	}
}

func TestPedanticPub(t *testing.T) {
	s := runPedanticServer()
	defer s.Shutdown()

	c := createClientConn(t, "127.0.0.1", PROTO_TEST_PORT)
	defer c.Close()

	send := sendCommand(t, c)
	expect := expectCommand(t, c)
	doConnect(t, c, false, true, false)

	// Ping should still be same
	send("PING\r\n")
	expect(pongRe)

	// Test malformed subjects for PUB
	// PUB subjects can not have wildcards
	// This will error in pedantic mode
	send("PUB foo.* 2\r\nok\r\n")
	expect(errRe)

	send("PUB foo.> 2\r\nok\r\n")
	expect(errRe)

	send("PUB foo. 2\r\nok\r\n")
	expect(errRe)

	send("PUB .foo 2\r\nok\r\n")
	expect(errRe)

	send("PUB foo..* 2\r\nok\r\n")
	expect(errRe)
}
