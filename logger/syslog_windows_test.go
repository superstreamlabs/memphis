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
package logger

import (
	"os/exec"
	"strings"
	"testing"

	"golang.org/x/sys/windows/svc/eventlog"
)

// Skips testing if we do not have privledges to run this test.
// This lets us skip the tests for general (non admin/system) users.
func checkPrivledges(t *testing.T) {
	src := "NATS-eventlog-testsource"
	defer eventlog.Remove(src)
	if err := eventlog.InstallAsEventCreate(src, eventlog.Info|eventlog.Error|eventlog.Warning); err != nil {
		if strings.Contains(err.Error(), "Access is denied") {
			// Skip this test because elevated privileges are required.
			t.SkipNow()
		}
		// let the tests report other types of errors
	}
}

// lastLogEntryContains reads the last entry (/c:1 /rd:true) written
// to the event log by the NATS-Server source, returning true if the
// passed text was found, false otherwise.
func lastLogEntryContains(t *testing.T, text string) bool {
	var output []byte
	var err error

	cmd := exec.Command("wevtutil.exe", "qe", "Application", "/q:*[System[Provider[@Name='NATS-Server']]]",
		"/rd:true", "/c:1")
	if output, err = cmd.Output(); err != nil {
		t.Fatalf("Unable to execute command: %v", err)
	}
	return strings.Contains(string(output), text)
}

// TestSysLogger tests event logging on windows
func TestSysLogger(t *testing.T) {
	checkPrivledges(t)
	logger := NewSysLogger(false, false)
	if logger.debug {
		t.Fatalf("Expected %t, received %t\n", false, logger.debug)
	}

	if logger.trace {
		t.Fatalf("Expected %t, received %t\n", false, logger.trace)
	}
	logger.Noticef("%s", "Noticef")
	if !lastLogEntryContains(t, "[NOTICE]: Noticef") {
		t.Fatalf("missing log entry")
	}

	logger.Errorf("%s", "Errorf")
	if !lastLogEntryContains(t, "[ERROR]: Errorf") {
		t.Fatalf("missing log entry")
	}

	logger.Tracef("%s", "Tracef")
	if lastLogEntryContains(t, "Tracef") {
		t.Fatalf("should not contain log entry")
	}

	logger.Debugf("%s", "Debugf")
	if lastLogEntryContains(t, "Debugf") {
		t.Fatalf("should not contain log entry")
	}
}

// TestSysLoggerWithDebugAndTrace tests event logging
func TestSysLoggerWithDebugAndTrace(t *testing.T) {
	checkPrivledges(t)
	logger := NewSysLogger(true, true)
	if !logger.debug {
		t.Fatalf("Expected %t, received %t\n", true, logger.debug)
	}

	if !logger.trace {
		t.Fatalf("Expected %t, received %t\n", true, logger.trace)
	}

	logger.Tracef("%s", "Tracef")
	if !lastLogEntryContains(t, "[TRACE]: Tracef") {
		t.Fatalf("missing log entry")
	}

	logger.Debugf("%s", "Debugf")
	if !lastLogEntryContains(t, "[DEBUG]: Debugf") {
		t.Fatalf("missing log entry")
	}
}

// TestSysLoggerWithDebugAndTrace tests remote event logging
func TestRemoteSysLoggerWithDebugAndTrace(t *testing.T) {
	checkPrivledges(t)
	logger := NewRemoteSysLogger("", true, true)
	if !logger.debug {
		t.Fatalf("Expected %t, received %t\n", true, logger.debug)
	}

	if !logger.trace {
		t.Fatalf("Expected %t, received %t\n", true, logger.trace)
	}
	logger.Tracef("NATS %s", "[TRACE]: Remote Noticef")
	if !lastLogEntryContains(t, "Remote Noticef") {
		t.Fatalf("missing log entry")
	}
}

func TestSysLoggerFatalf(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			if !lastLogEntryContains(t, "[FATAL]: Fatalf") {
				t.Fatalf("missing log entry")
			}
		}
	}()

	checkPrivledges(t)
	logger := NewSysLogger(true, true)
	logger.Fatalf("%s", "Fatalf")
	t.Fatalf("did not panic when expected to")
}
