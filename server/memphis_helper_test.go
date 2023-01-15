// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis-broker/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

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
			memphisMsgs, err := s.memphisGetMsgs("", mset.name(), 1, msgsToFetchNum, 1*time.Second, true)
			if err != nil {
				t.Fatalf("Unexpected error getting messages: %v", err)
			}
			if len(memphisMsgs) != 1 {
				t.Fatalf("Expected 1 message, got %d", len(memphisMsgs))
			}

			msgsToFetchNum = 2
			memphisMsgs, err = s.memphisGetMsgs("", mset.name(), 1, msgsToFetchNum, 1*time.Second, true)
			if err != nil {
				t.Fatalf("Unexpected error getting messages: %v", err)
			}
			if len(memphisMsgs) != 2 {
				t.Fatalf("Expected 2 message, got %d", len(memphisMsgs))
			}

			msgsToFetchNum = 3
			memphisMsgs, err = s.memphisGetMsgs("", mset.name(), 1, msgsToFetchNum, 1*time.Second, true)
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
