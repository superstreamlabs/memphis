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
package testhelper

// These routines need to be accessible in both the server and test
// directories, and tests importing a package don't get exported symbols from
// _test.go files in the imported package, so we put them here where they can
// be used freely.

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

type DummyLogger struct {
	sync.Mutex
	Msg     string
	AllMsgs []string
}

func (l *DummyLogger) CheckContent(t *testing.T, expectedStr string) {
	t.Helper()
	l.Lock()
	defer l.Unlock()
	if l.Msg != expectedStr {
		t.Fatalf("Expected log to be: %v, got %v", expectedStr, l.Msg)
	}
}

func (l *DummyLogger) aggregate() {
	if l.AllMsgs != nil {
		l.AllMsgs = append(l.AllMsgs, l.Msg)
	}
}

func (l *DummyLogger) Noticef(format string, v ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.Msg = fmt.Sprintf(format, v...)
	l.aggregate()
}
func (l *DummyLogger) Errorf(format string, v ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.Msg = fmt.Sprintf(format, v...)
	l.aggregate()
}
func (l *DummyLogger) Warnf(format string, v ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.Msg = fmt.Sprintf(format, v...)
	l.aggregate()
}
func (l *DummyLogger) Fatalf(format string, v ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.Msg = fmt.Sprintf(format, v...)
	l.aggregate()
}
func (l *DummyLogger) Debugf(format string, v ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.Msg = fmt.Sprintf(format, v...)
	l.aggregate()
}
func (l *DummyLogger) Tracef(format string, v ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.Msg = fmt.Sprintf(format, v...)
	l.aggregate()
}

// NewDummyLogger creates a dummy logger and allows to ask for logs to be
// retained instead of just keeping the most recent. Use retain to provide an
// initial size estimate on messages (not to provide a max capacity).
func NewDummyLogger(retain uint) *DummyLogger {
	l := &DummyLogger{}
	if retain > 0 {
		l.AllMsgs = make([]string, 0, retain)
	}
	return l
}

func (l *DummyLogger) Drain() {
	l.Lock()
	defer l.Unlock()
	if l.AllMsgs == nil {
		return
	}
	l.AllMsgs = make([]string, 0, len(l.AllMsgs))
}

func (l *DummyLogger) CheckForProhibited(t *testing.T, reason, needle string) {
	t.Helper()
	l.Lock()
	defer l.Unlock()

	if l.AllMsgs == nil {
		t.Fatal("DummyLogger.CheckForProhibited called without AllMsgs being collected")
	}

	// Collect _all_ matches, rather than have to re-test repeatedly.
	// This will particularly help with less deterministic tests with multiple matches.
	shouldFail := false
	for i := range l.AllMsgs {
		if strings.Contains(l.AllMsgs[i], needle) {
			t.Errorf("log contains %s: %v", reason, l.AllMsgs[i])
			shouldFail = true
		}
	}
	if shouldFail {
		t.FailNow()
	}
}
