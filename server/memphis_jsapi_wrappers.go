// Credit for The NATS.IO Authors
// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package server
package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"memphis-broker/models"
	"time"
)

const (
	// wrapper subject for JSApiTemplateCreate
	memphisJSApiStreamCreate = "$MEMPHIS.JS.API.STREAM.CREATE"
	// wrapper subject for JSApiTemplateDelete
	memphisJSApiStreamDelete = "$MEMPHIS.JS.API.STREAM.DELETE"
)

var wrapperMap = map[string]string{
	memphisJSApiStreamCreate: JSApiStreamCreate,
	memphisJSApiStreamDelete: JSApiStreamDelete,
}

func memphisFindJSAPIWrapperSubject(c *client, subject string) string {
	if c.memphisInfo.isNative ||
		c.kind != CLIENT {
		return subject
	}

	for wrapped, subj := range wrapperMap {
		if subjectIsSubsetMatch(subject, subj) {
			return wrapped
		}
	}

	return subject
}

func (s *Server) memphisJSApiWrapStreamCreate(sub *subscription, c *client, acc *Account, subject, reply string, rmsg []byte) {
	var resp = JSApiStreamCreateResponse{ApiResponse: ApiResponse{Type: JSApiStreamCreateResponseType}}

	var cfg StreamConfig
	ci, acc, _, msg, err := s.getRequestInfo(c, rmsg)
	if err != nil {
		s.Warnf(badAPIRequestT, msg)
		return
	}

	if err := json.Unmarshal(msg, &cfg); err != nil {
		resp.Error = NewJSInvalidJSONError()
		s.sendAPIErrResponse(ci, acc, subject, reply, string(msg), s.jsonResponse(&resp))
		return
	}

	if cfg.Retention != LimitsPolicy {
		resp.Error = NewJSStreamCreateError(errors.New("The only supported retention type is limits"))
		s.sendAPIErrResponse(ci, acc, subject, reply, string(msg), s.jsonResponse(&resp))
		return
	}

	createStreamFunc := func() error {
		if !s.jsStreamCreateRequestIntern(sub, c, acc, subject, reply, rmsg) {
			return errors.New("Stream creation failed")
		}
		return nil
	}
	if s.JetStreamIsClustered() && s.JetStreamIsLeader() {
		s.Warnf(fmt.Sprintf("create stream %s: 111", cfg.Name))
		quit := make(chan bool)
		resultCh := make(chan *StreamInfo)
		go func(quitCh chan bool, resultChan chan *StreamInfo) {
			ticker := time.NewTicker(500 * time.Millisecond)
			for {
				select {
				case <-ticker.C:
					info, err := s.memphisStreamInfo(cfg.Name)
					if err == nil {
						resultChan <- info
						return
					}
				case <-quitCh:
					return
				}
			}
		}(quit, resultCh)

		var result *StreamInfo
		select {
		case result = <-resultCh:
			s.Warnf(fmt.Sprintf("create stream %s: 222", cfg.Name))
			break
		case <-time.After(5 * time.Second):
			s.Warnf(fmt.Sprintf("create stream %s: 333", cfg.Name))
			quit <- true
		}

		if result == nil {
			return
		}

		createStreamFunc = func() error {
			return nil
		}
	}
	s.Warnf(fmt.Sprintf("create stream %s: 444", cfg.Name))
	var storageType string
	if cfg.Storage == MemoryStorage {
		storageType = "memory"
	} else if cfg.Storage == FileStorage {
		storageType = "file"
	}

	csr := createStationRequest{
		StationName:       cfg.Name,
		SchemaName:        "",
		RetentionType:     "",
		RetentionValue:    0,
		StorageType:       storageType,
		Replicas:          cfg.Replicas,
		DedupEnabled:      true,
		DedupWindowMillis: 0,
		IdempotencyWindow: int64(cfg.Duplicates.Milliseconds()),
		DlsConfiguration: models.DlsConfiguration{
			Poison:      true,
			Schemaverse: false,
		},
	}

	s.createStationDirectIntern(c, reply, &csr, createStreamFunc)
}

func (s *Server) memphisJSApiWrapStreamDelete(sub *subscription, c *client, acc *Account, subject, reply string, rmsg []byte) {
	removeStreamFunc := func() error {
		if !s.jsStreamDeleteRequestIntern(sub, c, acc, subject, reply, rmsg) {
			return errors.New("Stream removal failed")
		}
		return nil
	}

	if s.JetStreamIsClustered() && s.JetStreamIsLeader() {
		s.Warnf(fmt.Sprintf("delete stream %s: 111", streamNameFromSubject(subject)))
		quit := make(chan bool)
		resultCh := make(chan error)
		go func(quitCh chan bool, resultChan chan error) {
			ticker := time.NewTicker(500 * time.Millisecond)
			for {
				select {
				case <-ticker.C:
					_, err := s.memphisStreamInfo(streamNameFromSubject(subject))
					if IsNatsErr(err, JSStreamNotFoundErr) {
						resultChan <- err
						return
					}
				case <-quitCh:
					return
				}
			}
		}(quit, resultCh)

		var err error
		select {
		case err = <-resultCh:
			s.Warnf(fmt.Sprintf("delete stream %s: 222", streamNameFromSubject(subject)))
			break
		case <-time.After(5 * time.Second):
			s.Warnf(fmt.Sprintf("delete stream %s: 333", streamNameFromSubject(subject)))
			quit <- true
		}

		if err == nil {
			return
		}

		removeStreamFunc = func() error {
			return nil
		}
	}
	s.Warnf(fmt.Sprintf("delete stream %s: 444", streamNameFromSubject(subject)))
	dsr := destroyStationRequest{StationName: streamNameFromSubject(subject)}
	s.removeStationDirectIntern(c, reply, &dsr, removeStreamFunc)
}
