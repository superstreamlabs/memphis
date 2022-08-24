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
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/windows/svc/eventlog"
)

var natsEventSource = "NATS-Server"

// SetSyslogName sets the name to use for the system log event source
func SetSyslogName(name string) {
	natsEventSource = name
}

// SysLogger logs to the windows event logger
type SysLogger struct {
	writer *eventlog.Log
	debug  bool
	trace  bool
}

// NewSysLogger creates a log using the windows event logger
func NewSysLogger(debug, trace bool) *SysLogger {
	if err := eventlog.InstallAsEventCreate(natsEventSource, eventlog.Info|eventlog.Error|eventlog.Warning); err != nil {
		if !strings.Contains(err.Error(), "registry key already exists") {
			panic(fmt.Sprintf("could not access event log: %v", err))
		}
	}

	w, err := eventlog.Open(natsEventSource)
	if err != nil {
		panic(fmt.Sprintf("could not open event log: %v", err))
	}

	return &SysLogger{
		writer: w,
		debug:  debug,
		trace:  trace,
	}
}

// NewRemoteSysLogger creates a remote event logger
func NewRemoteSysLogger(fqn string, debug, trace bool) *SysLogger {
	w, err := eventlog.OpenRemote(fqn, natsEventSource)
	if err != nil {
		panic(fmt.Sprintf("could not open event log: %v", err))
	}

	return &SysLogger{
		writer: w,
		debug:  debug,
		trace:  trace,
	}
}

func formatMsg(tag, format string, v ...interface{}) string {
	orig := fmt.Sprintf(format, v...)
	return fmt.Sprintf("pid[%d][%s]: %s", os.Getpid(), tag, orig)
}

// Noticef logs a notice statement
func (l *SysLogger) Noticef(format string, v ...interface{}) {
	l.writer.Info(1, formatMsg("NOTICE", format, v...))
}

// Warnf logs a warning statement
func (l *SysLogger) Warnf(format string, v ...interface{}) {
	l.writer.Info(1, formatMsg("WARN", format, v...))
}

// Fatalf logs a fatal error
func (l *SysLogger) Fatalf(format string, v ...interface{}) {
	msg := formatMsg("FATAL", format, v...)
	l.writer.Error(5, msg)
	panic(msg)
}

// Errorf logs an error statement
func (l *SysLogger) Errorf(format string, v ...interface{}) {
	l.writer.Error(2, formatMsg("ERROR", format, v...))
}

// Debugf logs a debug statement
func (l *SysLogger) Debugf(format string, v ...interface{}) {
	if l.debug {
		l.writer.Info(3, formatMsg("DEBUG", format, v...))
	}
}

// Tracef logs a trace statement
func (l *SysLogger) Tracef(format string, v ...interface{}) {
	if l.trace {
		l.writer.Info(4, formatMsg("TRACE", format, v...))
	}
}
