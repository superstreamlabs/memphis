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
	"reflect"
	"testing"
)

func TestRBAppendAndLenAndTotal(t *testing.T) {
	rb := newClosedRingBuffer(10)
	for i := 0; i < 5; i++ {
		rb.append(&closedClient{})
	}
	if rbl := rb.len(); rbl != 5 {
		t.Fatalf("Expected len of 5, got %d", rbl)
	}
	if rbt := rb.totalConns(); rbt != 5 {
		t.Fatalf("Expected total of 5, got %d", rbt)
	}
	for i := 0; i < 25; i++ {
		rb.append(&closedClient{})
	}
	if rbl := rb.len(); rbl != 10 {
		t.Fatalf("Expected len of 10, got %d", rbl)
	}
	if rbt := rb.totalConns(); rbt != 30 {
		t.Fatalf("Expected total of 30, got %d", rbt)
	}
}

func (cc *closedClient) String() string {
	return cc.user
}

func TestRBclosedClients(t *testing.T) {
	rb := newClosedRingBuffer(10)

	var ui int
	addConn := func() {
		ui++
		rb.append(&closedClient{user: fmt.Sprintf("%d", ui)})
	}

	max := 100
	master := make([]*closedClient, 0, max)
	for i := 1; i <= max; i++ {
		master = append(master, &closedClient{user: fmt.Sprintf("%d", i)})
	}

	testList := func(i int) {
		ccs := rb.closedClients()
		start := int(rb.totalConns()) - len(ccs)
		ms := master[start : start+len(ccs)]
		if !reflect.DeepEqual(ccs, ms) {
			t.Fatalf("test %d: List result did not match master: %+v vs %+v", i, ccs, ms)
		}
	}

	for i := 0; i < max; i++ {
		addConn()
		testList(i)
	}
}
