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

// We wrap to hold onto optional items for /connz.
type closedClient struct {
	ConnInfo
	subs []SubDetail
	user string
	acc  string
}

// Fixed sized ringbuffer for closed connections.
type closedRingBuffer struct {
	total uint64
	conns []*closedClient
}

// Create a new ring buffer with at most max items.
func newClosedRingBuffer(max int) *closedRingBuffer {
	rb := &closedRingBuffer{}
	rb.conns = make([]*closedClient, max)
	return rb
}

// Adds in a new closed connection. If there is no more room,
// remove the oldest.
func (rb *closedRingBuffer) append(cc *closedClient) {
	rb.conns[rb.next()] = cc
	rb.total++
}

func (rb *closedRingBuffer) next() int {
	return int(rb.total % uint64(cap(rb.conns)))
}

func (rb *closedRingBuffer) len() int {
	if rb.total > uint64(cap(rb.conns)) {
		return cap(rb.conns)
	}
	return int(rb.total)
}

func (rb *closedRingBuffer) totalConns() uint64 {
	return rb.total
}

// This will return a sorted copy of the list which recipient can
// modify. If the contents of the client itself need to be modified,
// meaning swapping in any optional items, a copy should be made. We
// could introduce a new lock and hold that but since we return this
// list inside monitor which allows programatic access, we do not
// know when it would be done.
func (rb *closedRingBuffer) closedClients() []*closedClient {
	dup := make([]*closedClient, rb.len())
	head := rb.next()
	if rb.total <= uint64(cap(rb.conns)) || head == 0 {
		copy(dup, rb.conns[:rb.len()])
	} else {
		fp := rb.conns[head:]
		sp := rb.conns[:head]
		copy(dup, fp)
		copy(dup[len(fp):], sp)
	}
	return dup
}
