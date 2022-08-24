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
	"strconv"
	"sync"
)

type outMsg struct {
	subj string
	rply string
	hdr  []byte
	msg  []byte
}

type sendq struct {
	mu sync.Mutex
	q  *ipQueue // of *outMsg
	s  *Server
}

func (s *Server) newSendQ() *sendq {
	sq := &sendq{s: s, q: s.newIPQueue("SendQ")}
	s.startGoRoutine(sq.internalLoop)
	return sq
}

func (sq *sendq) internalLoop() {
	sq.mu.Lock()
	s, q := sq.s, sq.q
	sq.mu.Unlock()

	defer s.grWG.Done()

	c := s.createInternalSystemClient()
	c.registerWithAccount(s.SystemAccount())
	c.noIcb = true

	defer c.closeConnection(ClientClosed)

	for s.isRunning() {
		select {
		case <-s.quitCh:
			return
		case <-q.ch:
			pms := q.pop()
			for _, pmi := range pms {
				pm := pmi.(*outMsg)
				c.pa.subject = []byte(pm.subj)
				c.pa.size = len(pm.msg) + len(pm.hdr)
				c.pa.szb = []byte(strconv.Itoa(c.pa.size))
				c.pa.reply = []byte(pm.rply)
				var msg []byte
				if len(pm.hdr) > 0 {
					c.pa.hdr = len(pm.hdr)
					c.pa.hdb = []byte(strconv.Itoa(c.pa.hdr))
					msg = append(pm.hdr, pm.msg...)
					msg = append(msg, _CRLF_...)
				} else {
					c.pa.hdr = -1
					c.pa.hdb = nil
					msg = append(pm.msg, _CRLF_...)
				}
				c.processInboundClientMsg(msg)
				c.pa.szb = nil
			}
			// TODO: should this be in the for-loop instead?
			c.flushClients(0)
			q.recycle(&pms)
		}
	}
}

func (sq *sendq) send(subj, rply string, hdr, msg []byte) {
	out := &outMsg{subj, rply, nil, nil}
	// We will copy these for now.
	if len(hdr) > 0 {
		hdr = copyBytes(hdr)
		out.hdr = hdr
	}
	if len(msg) > 0 {
		msg = copyBytes(msg)
		out.msg = msg
	}
	sq.q.push(out)
}
