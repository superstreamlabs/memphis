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
	"crypto/tls"
	"fmt"
	"net"
	"testing"
	"time"

	"memphis-broker/server"
)

const (
	PING_TEST_PORT = 9972
	PING_INTERVAL  = 50 * time.Millisecond
	PING_MAX       = 2
)

func runPingServer() *server.Server {
	opts := DefaultTestOptions
	opts.Port = PING_TEST_PORT
	opts.PingInterval = PING_INTERVAL
	opts.MaxPingsOut = PING_MAX
	return RunServer(&opts)
}

func TestPingSentToTLSConnection(t *testing.T) {
	opts := DefaultTestOptions
	opts.Port = PING_TEST_PORT
	opts.PingInterval = PING_INTERVAL
	opts.MaxPingsOut = PING_MAX
	opts.TLSCert = "configs/certs/server-cert.pem"
	opts.TLSKey = "configs/certs/server-key.pem"
	opts.TLSCaCert = "configs/certs/ca.pem"

	tc := server.TLSConfigOpts{}
	tc.CertFile = opts.TLSCert
	tc.KeyFile = opts.TLSKey
	tc.CaFile = opts.TLSCaCert

	opts.TLSConfig, _ = server.GenTLSConfig(&tc)
	opts.TLSTimeout = 5
	s := RunServer(&opts)
	defer s.Shutdown()

	c := createClientConn(t, "127.0.0.1", PING_TEST_PORT)
	defer c.Close()

	checkInfoMsg(t, c)
	c = tls.Client(c, &tls.Config{InsecureSkipVerify: true})
	tlsConn := c.(*tls.Conn)
	tlsConn.Handshake()

	cs := fmt.Sprintf("CONNECT {\"verbose\":%v,\"pedantic\":%v,\"tls_required\":%v}\r\n", false, false, true)
	sendProto(t, c, cs)

	expect := expectCommand(t, c)

	// Expect the max to be delivered correctly..
	for i := 0; i < PING_MAX; i++ {
		time.Sleep(PING_INTERVAL / 2)
		expect(pingRe)
	}

	// We should get an error from the server
	time.Sleep(PING_INTERVAL)
	expect(errRe)

	// Server should close the connection at this point..
	time.Sleep(PING_INTERVAL)
	c.SetWriteDeadline(time.Now().Add(PING_INTERVAL))

	var err error
	for {
		_, err = c.Write([]byte("PING\r\n"))
		if err != nil {
			break
		}
	}
	c.SetWriteDeadline(time.Time{})

	if err == nil {
		t.Fatal("No error: Expected to have connection closed")
	}
	if ne, ok := err.(net.Error); ok && ne.Timeout() {
		t.Fatal("timeout: Expected to have connection closed")
	}
}

func TestPingInterval(t *testing.T) {
	s := runPingServer()
	defer s.Shutdown()

	c := createClientConn(t, "127.0.0.1", PING_TEST_PORT)
	defer c.Close()

	doConnect(t, c, false, false, false)

	expect := expectCommand(t, c)

	// Expect the max to be delivered correctly..
	for i := 0; i < PING_MAX; i++ {
		time.Sleep(PING_INTERVAL / 2)
		expect(pingRe)
	}

	// We should get an error from the server
	time.Sleep(PING_INTERVAL)
	expect(errRe)

	// Server should close the connection at this point..
	time.Sleep(PING_INTERVAL)
	c.SetWriteDeadline(time.Now().Add(PING_INTERVAL))

	var err error
	for {
		_, err = c.Write([]byte("PING\r\n"))
		if err != nil {
			break
		}
	}
	c.SetWriteDeadline(time.Time{})

	if err == nil {
		t.Fatal("No error: Expected to have connection closed")
	}
	if ne, ok := err.(net.Error); ok && ne.Timeout() {
		t.Fatal("timeout: Expected to have connection closed")
	}
}

func TestUnpromptedPong(t *testing.T) {
	s := runPingServer()
	defer s.Shutdown()

	c := createClientConn(t, "127.0.0.1", PING_TEST_PORT)
	defer c.Close()

	doConnect(t, c, false, false, false)

	expect := expectCommand(t, c)

	// Send lots of PONGs in a row...
	for i := 0; i < 100; i++ {
		c.Write([]byte("PONG\r\n"))
	}

	// The server should still send the max number of PINGs and then
	// close the connection.
	for i := 0; i < PING_MAX; i++ {
		time.Sleep(PING_INTERVAL / 2)
		expect(pingRe)
	}

	// We should get an error from the server
	time.Sleep(PING_INTERVAL)
	expect(errRe)

	// Server should close the connection at this point..
	c.SetWriteDeadline(time.Now().Add(PING_INTERVAL))
	var err error
	for {
		_, err = c.Write([]byte("PING\r\n"))
		if err != nil {
			break
		}
	}
	c.SetWriteDeadline(time.Time{})

	if err == nil {
		t.Fatal("No error: Expected to have connection closed")
	}
	if ne, ok := err.(net.Error); ok && ne.Timeout() {
		t.Fatal("timeout: Expected to have connection closed")
	}
}

func TestPingSuppresion(t *testing.T) {
	pingInterval := 100 * time.Millisecond
	highWater := 130 * time.Millisecond
	opts := DefaultTestOptions
	opts.Port = PING_TEST_PORT
	opts.PingInterval = pingInterval
	opts.DisableShortFirstPing = true

	s := RunServer(&opts)
	defer s.Shutdown()

	c := createClientConn(t, "127.0.0.1", PING_TEST_PORT)
	defer c.Close()

	connectTime := time.Now()

	send, expect := setupConn(t, c)

	expect(pingRe)
	pingTime := time.Since(connectTime)
	send("PONG\r\n")

	// Should be > 100 but less then 120(ish)
	if pingTime < pingInterval {
		t.Fatalf("pingTime too low: %v", pingTime)
	}
	// +5 is just for fudging in case things are slow in the testing system.
	if pingTime > highWater {
		t.Fatalf("pingTime too high: %v", pingTime)
	}

	time.Sleep(pingInterval / 2)

	// Sending a PING should suppress.
	send("PING\r\n")
	expect(pongRe)

	// This will wait for the time period where a PING should have fired
	// and been delivered. We expect nothing here since it should be suppressed.
	expectNothingTimeout(t, c, time.Now().Add(100*time.Millisecond))
}
