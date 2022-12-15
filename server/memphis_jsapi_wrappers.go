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

import "encoding/json"

const (
	// wrapper subject for JSApiTemplateCreate
	memphisJSApiStreamCreate = "$MEMPHIS.JS.API.STREAM.CREATE"
	// wrapper subject for JSApiTemplateDelete
	memphisJSApiStreamDelete = "$MEMPHIS.JS.API.STREAM.DELETE"
)

var wrapperMap = map[string]string{
	memphisJSApiStreamCreate: JSApiStreamCreate,
	memphisJSApiStreamDelete: JSApiTemplateDelete,
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
		resp.Error = NewJSInvalidJSONError()
		s.sendAPIErrResponse(ci, acc, subject, reply, string(msg), s.jsonResponse(&resp))
		return
	}

	streamCreated := s.jsStreamCreateRequestIntern(sub, c, acc, subject, reply, rmsg)
	if !streamCreated {
		return
	}

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
		IdempotencyWindow: int(cfg.Duplicates.Milliseconds()),
	}

	s.createStationDirectIntern(c, reply, &csr, false)
}

func (s *Server) memphisJSApiWrapStreamDelete(sub *subscription, c *client, acc *Account, subject, reply string, rmsg []byte) {

	streamDeleted := s.jsStreamDeleteRequestIntern(sub, c, acc, subject, reply, rmsg)

	if !streamDeleted {
		return
	}

	dsr := destroyStationRequest{StationName: streamNameFromSubject(subject)}
	s.removeStationDirectIntern(c, reply, &dsr)
}
