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

//go:build !skip_js_tests
// +build !skip_js_tests

package server

import (
	"testing"
	"time"
)

func TestMemphisGetMsgs(t *testing.T) {
	cases := []struct {
		name    string
		mconfig *StreamConfig
	}{
		{name: "MemoryStore",
			mconfig: &StreamConfig{
				Name:      "foo",
				Retention: LimitsPolicy,
				MaxAge:    time.Hour,
				Storage:   MemoryStorage,
				Replicas:  1,
			}},
		{name: "FileStore",
			mconfig: &StreamConfig{
				Name:      "foo",
				Retention: LimitsPolicy,
				MaxAge:    time.Hour,
				Storage:   FileStorage,
				Replicas:  1,
			}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := RunBasicJetStreamServer()
			defer s.Shutdown()

			if config := s.JetStreamConfig(); config != nil {
				defer removeDir(t, config.StoreDir)
			}

			mset, err := s.GlobalAccount().addStream(c.mconfig)
			if err != nil {
				t.Fatalf("Unexpected error adding stream: %v", err)
			}
			defer mset.delete()

			nc := clientConnectToServer(t, s)
			defer nc.Close()

			nc.Publish("foo", []byte("Hello World!"))
			nc.Flush()

			state := mset.state()
			if state.Msgs != 1 {
				t.Fatalf("Expected 1 message, got %d", state.Msgs)
			}
			if state.Bytes == 0 {
				t.Fatalf("Expected non-zero bytes")
			}

			nc.Publish("foo", []byte("Hello World Again!"))
			nc.Flush()

			state = mset.state()
			if state.Msgs != 2 {
				t.Fatalf("Expected 2 messages, got %d", state.Msgs)
			}

			msgsToFetchNum := 1
			memphisMsgs, err := s.memphisGetMsgs("foo", mset.name(), 1, msgsToFetchNum, 1*time.Second)
			if err != nil {
				t.Fatalf("Unexpected error getting messages: %v", err)
			}
			if len(memphisMsgs) != 1 {
				t.Fatalf("Expected 1 message, got %d", len(memphisMsgs))
			}

			msgsToFetchNum = 2
			memphisMsgs, err = s.memphisGetMsgs("foo", mset.name(), 1, msgsToFetchNum, 1*time.Second)
			if err != nil {
				t.Fatalf("Unexpected error getting messages: %v", err)
			}
			if len(memphisMsgs) != 2 {
				t.Fatalf("Expected 2 message, got %d", len(memphisMsgs))
			}

			msgsToFetchNum = 3
			memphisMsgs, err = s.memphisGetMsgs("foo", mset.name(), 1, msgsToFetchNum, 1*time.Second)
			if err != ErrStoreEOF {
				t.Fatalf("Unexpected error getting messages: %v", err)
			}

			if err := mset.delete(); err != nil {
				t.Fatalf("Got an error deleting the stream: %v", err)
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	validName := "abc123._-"
	invalidName := "$isWhatINeed(hey,hey)"

	if validateName(validName, "some type") != nil {
		t.Error()
	}

	if validateName(invalidName, "some type") == nil {
		t.Error()
	}
}
