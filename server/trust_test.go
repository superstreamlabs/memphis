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
	"strings"
	"testing"
)

const (
	t1 = "OBYEOZQ46VZMFMNETBAW2H6VGDSOBLP67VUEZJ5LPR3PIZBWWRIY4UI4"
	t2 = "OAHC7NGAHG3YVPTD6QOUFZGPM2OMU6EOS67O2VHBUOA6BJLPTWFHGLKU"
)

func TestStampedTrustedKeys(t *testing.T) {
	opts := DefaultOptions()
	defer func() { trustedKeys = "" }()

	// Set this to a bad key. We require valid operator public keys.
	trustedKeys = "bad"
	if s := New(opts); s != nil {
		s.Shutdown()
		t.Fatalf("Expected a bad trustedKeys to return nil server")
	}

	trustedKeys = t1
	s := New(opts)
	if s == nil {
		t.Fatalf("Expected non-nil server")
	}
	if len(s.trustedKeys) != 1 || s.trustedKeys[0] != t1 {
		t.Fatalf("Trusted Nkeys not setup properly")
	}
	trustedKeys = strings.Join([]string{t1, t2}, " ")
	if s = New(opts); s == nil {
		t.Fatalf("Expected non-nil server")
	}
	if len(s.trustedKeys) != 2 || s.trustedKeys[0] != t1 || s.trustedKeys[1] != t2 {
		t.Fatalf("Trusted Nkeys not setup properly")
	}

	opts.TrustedKeys = []string{"OVERRIDE ME"}
	if s = New(opts); s != nil {
		t.Fatalf("Expected opts.TrustedKeys to return nil server")
	}
}

func TestTrustedKeysOptions(t *testing.T) {
	trustedKeys = ""
	opts := DefaultOptions()
	opts.TrustedKeys = []string{"bad"}
	if s := New(opts); s != nil {
		s.Shutdown()
		t.Fatalf("Expected a bad opts.TrustedKeys to return nil server")
	}
	opts.TrustedKeys = []string{t1}
	s := New(opts)
	if s == nil {
		t.Fatalf("Expected non-nil server")
	}
	if len(s.trustedKeys) != 1 || s.trustedKeys[0] != t1 {
		t.Fatalf("Trusted Nkeys not setup properly via options")
	}
	opts.TrustedKeys = []string{t1, t2}
	if s = New(opts); s == nil {
		t.Fatalf("Expected non-nil server")
	}
	if len(s.trustedKeys) != 2 || s.trustedKeys[0] != t1 || s.trustedKeys[1] != t2 {
		t.Fatalf("Trusted Nkeys not setup properly via options")
	}
}

func TestTrustConfigOption(t *testing.T) {
	confFileName := createConfFile(t, []byte(fmt.Sprintf("trusted = %q", t1)))
	defer removeFile(t, confFileName)
	opts, err := ProcessConfigFile(confFileName)
	if err != nil {
		t.Fatalf("Error parsing config: %v", err)
	}
	if l := len(opts.TrustedKeys); l != 1 {
		t.Fatalf("Expected 1 trusted key, got %d", l)
	}
	if opts.TrustedKeys[0] != t1 {
		t.Fatalf("Expected trusted key to be %q, got %q", t1, opts.TrustedKeys[0])
	}

	confFileName = createConfFile(t, []byte(fmt.Sprintf("trusted = [%q, %q]", t1, t2)))
	defer removeFile(t, confFileName)
	opts, err = ProcessConfigFile(confFileName)
	if err != nil {
		t.Fatalf("Error parsing config: %v", err)
	}
	if l := len(opts.TrustedKeys); l != 2 {
		t.Fatalf("Expected 2 trusted key, got %d", l)
	}
	if opts.TrustedKeys[0] != t1 {
		t.Fatalf("Expected trusted key to be %q, got %q", t1, opts.TrustedKeys[0])
	}
	if opts.TrustedKeys[1] != t2 {
		t.Fatalf("Expected trusted key to be %q, got %q", t2, opts.TrustedKeys[1])
	}

	// Now do a bad one.
	confFileName = createConfFile(t, []byte(fmt.Sprintf("trusted = [%q, %q]", t1, "bad")))
	defer removeFile(t, confFileName)
	_, err = ProcessConfigFile(confFileName)
	if err == nil {
		t.Fatalf("Expected an error parsing trust keys with a bad key")
	}
}
