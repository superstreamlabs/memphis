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
	"encoding/base64"
)

// Raw length of the nonce challenge
const (
	nonceRawLen = 11
	nonceLen    = 15 // base64.RawURLEncoding.EncodedLen(nonceRawLen)
)

// NonceRequired tells us if we should send a nonce.
func (s *Server) NonceRequired() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.nonceRequired()
}

// nonceRequired tells us if we should send a nonce.
// Lock should be held on entry.
func (s *Server) nonceRequired() bool {
	return s.opts.AlwaysEnableNonce || len(s.nkeys) > 0 || s.trustedKeys != nil
}

// Generate a nonce for INFO challenge.
// Assumes server lock is held
func (s *Server) generateNonce(n []byte) {
	var raw [nonceRawLen]byte
	data := raw[:]
	s.prand.Read(data)
	base64.RawURLEncoding.Encode(n, data)
}
