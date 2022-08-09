// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
			memphisMsgs, err := s.MemphisGetMsgs("foo", mset.name(), 1, msgsToFetchNum, 1*time.Second)
			if err != nil {
				t.Fatalf("Unexpected error getting messages: %v", err)
			}
			if len(memphisMsgs) != 1 {
				t.Fatalf("Expected 1 message, got %d", len(memphisMsgs))
			}

			msgsToFetchNum = 2
			memphisMsgs, err = s.MemphisGetMsgs("foo", mset.name(), 1, msgsToFetchNum, 1*time.Second)
			if err != nil {
				t.Fatalf("Unexpected error getting messages: %v", err)
			}
			if len(memphisMsgs) != 2 {
				t.Fatalf("Expected 2 message, got %d", len(memphisMsgs))
			}

			msgsToFetchNum = 3
			memphisMsgs, err = s.MemphisGetMsgs("foo", mset.name(), 1, msgsToFetchNum, 1*time.Second)
			if err != ErrStoreEOF {
				t.Fatalf("Unexpected error getting messages: %v", err)
			}

			if err := mset.delete(); err != nil {
				t.Fatalf("Got an error deleting the stream: %v", err)
			}
		})
	}
}
