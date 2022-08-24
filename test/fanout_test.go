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

//go:build !race && !skipnoracetests
// +build !race,!skipnoracetests

package test

import (
	"fmt"
	"sync"
	"testing"

	"memphis-broker/server"

	"github.com/nats-io/nats.go"
)

// IMPORTANT: Tests in this file are not executed when running with the -race flag.
//            The test name should be prefixed with TestNoRace so we can run only
//            those tests: go test -run=TestNoRace ...

// As we look to improve high fanout situations make sure we
// have a test that checks ordering for all subscriptions from a single subscriber.
func TestNoRaceHighFanoutOrdering(t *testing.T) {
	opts := &server.Options{Host: "127.0.0.1", Port: server.RANDOM_PORT}

	s := RunServer(opts)
	defer s.Shutdown()

	url := fmt.Sprintf("nats://%s", s.Addr())

	const (
		nconns = 100
		nsubs  = 100
		npubs  = 500
	)

	// make unique
	subj := nats.NewInbox()

	var wg sync.WaitGroup
	wg.Add(nconns * nsubs)

	for i := 0; i < nconns; i++ {
		nc, err := nats.Connect(url)
		if err != nil {
			t.Fatalf("Expected a successful connect on %d, got %v\n", i, err)
		}

		nc.SetErrorHandler(func(c *nats.Conn, s *nats.Subscription, e error) {
			t.Fatalf("Got an error %v for %+v\n", s, err)
		})

		ec, _ := nats.NewEncodedConn(nc, nats.DEFAULT_ENCODER)

		for y := 0; y < nsubs; y++ {
			expected := 0
			ec.Subscribe(subj, func(n int) {
				if n != expected {
					t.Fatalf("Expected %d but received %d\n", expected, n)
				}
				expected++
				if expected >= npubs {
					wg.Done()
				}
			})
		}
		ec.Flush()
		defer ec.Close()
	}

	nc, _ := nats.Connect(url)
	ec, _ := nats.NewEncodedConn(nc, nats.DEFAULT_ENCODER)

	for i := 0; i < npubs; i++ {
		ec.Publish(subj, i)
	}
	defer ec.Close()

	wg.Wait()
}

func TestNoRaceRouteFormTimeWithHighSubscriptions(t *testing.T) {
	srvA, optsA := RunServerWithConfig("./configs/srv_a.conf")
	defer srvA.Shutdown()

	clientA := createClientConn(t, optsA.Host, optsA.Port)
	defer clientA.Close()

	sendA, expectA := setupConn(t, clientA)

	// Now add lots of subscriptions. These will need to be forwarded
	// to new routes when they are added.
	subsTotal := 100000
	for i := 0; i < subsTotal; i++ {
		subject := fmt.Sprintf("FOO.BAR.BAZ.%d", i)
		sendA(fmt.Sprintf("SUB %s %d\r\n", subject, i))
	}
	sendA("PING\r\n")
	expectA(pongRe)

	srvB, _ := RunServerWithConfig("./configs/srv_b.conf")
	defer srvB.Shutdown()

	checkClusterFormed(t, srvA, srvB)

	// Now wait for all subscriptions to be processed.
	if err := checkExpectedSubs(subsTotal, srvB); err != nil {
		// Make sure we are not a slow consumer
		// Check for slow consumer status
		if srvA.NumSlowConsumers() > 0 {
			t.Fatal("Did not receive all subscriptions due to slow consumer")
		} else {
			t.Fatalf("%v", err)
		}
	}
	// Just double check the slow consumer status.
	if srvA.NumSlowConsumers() > 0 {
		t.Fatalf("Received a slow consumer notification: %d", srvA.NumSlowConsumers())
	}
}
