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
)

func TestVerbosePing(t *testing.T) {
	s := runProtoServer()
	defer s.Shutdown()

	c := createClientConn(t, "127.0.0.1", PROTO_TEST_PORT)
	defer c.Close()

	doConnect(t, c, true, false, false)

	send := sendCommand(t, c)
	expect := expectCommand(t, c)

	expect(okRe)

	// Ping should still be same
	send("PING\r\n")
	expect(pongRe)
}

func TestVerboseConnect(t *testing.T) {
	s := runProtoServer()
	defer s.Shutdown()

	c := createClientConn(t, "127.0.0.1", PROTO_TEST_PORT)
	defer c.Close()

	doConnect(t, c, true, false, false)

	send := sendCommand(t, c)
	expect := expectCommand(t, c)

	expect(okRe)

	// Connect
	send("CONNECT {\"verbose\":true,\"pedantic\":true,\"tls_required\":false}\r\n")
	expect(okRe)
}

func TestVerbosePubSub(t *testing.T) {
	s := runProtoServer()
	defer s.Shutdown()

	c := createClientConn(t, "127.0.0.1", PROTO_TEST_PORT)
	defer c.Close()

	doConnect(t, c, true, false, false)
	send := sendCommand(t, c)
	expect := expectCommand(t, c)

	expect(okRe)

	// Pub
	send("PUB foo 2\r\nok\r\n")
	expect(okRe)

	// Sub
	send("SUB foo 1\r\n")
	expect(okRe)

	// UnSub
	send("UNSUB 1\r\n")
	expect(okRe)
}
