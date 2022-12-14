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
	s.Errorf("Yay we wrapped it!")
	s.jsStreamCreateRequest(sub, c, acc, subject, reply, rmsg)
}

func (s *Server) memphisJSApiWrapStreamDelete(sub *subscription, c *client, acc *Account, subject, reply string, rmsg []byte) {
	s.Errorf("Yay we wrapped it!")
	s.jsStreamDeleteRequest(sub, c, acc, subject, reply, rmsg)
}
