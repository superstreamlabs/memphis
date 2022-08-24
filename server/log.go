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
	"io"
	"os"
	"sync/atomic"
	"time"

	srvlog "memphis-broker/logger"
)

// Logger interface of the NATS Server
type Logger interface {

	// Log a notice statement
	Noticef(format string, v ...interface{})

	// Log a warning statement
	Warnf(format string, v ...interface{})

	// Log a fatal error
	Fatalf(format string, v ...interface{})

	// Log an error
	Errorf(format string, v ...interface{})

	// Log a debug statement
	Debugf(format string, v ...interface{})

	// Log a trace statement
	Tracef(format string, v ...interface{})
}

// ConfigureLogger configures and sets the logger for the server.
func (s *Server) ConfigureLogger() {
	var (
		log Logger

		// Snapshot server options.
		opts = s.getOpts()
	)

	if opts.NoLog {
		return
	}

	syslog := opts.Syslog
	if isWindowsService() && opts.LogFile == "" {
		// Enable syslog if no log file is specified and we're running as a
		// Windows service so that logs are written to the Windows event log.
		syslog = true
	}

	if opts.LogFile != "" {
		log = srvlog.NewFileLogger(opts.LogFile, opts.Logtime, opts.Debug, opts.Trace, true)
		if opts.LogSizeLimit > 0 {
			if l, ok := log.(*srvlog.Logger); ok {
				l.SetSizeLimit(opts.LogSizeLimit)
			}
		}
	} else if opts.RemoteSyslog != "" {
		log = srvlog.NewRemoteSysLogger(opts.RemoteSyslog, opts.Debug, opts.Trace)
	} else if syslog {
		log = srvlog.NewSysLogger(opts.Debug, opts.Trace)
	} else {
		colors := true
		// Check to see if stderr is being redirected and if so turn off color
		// Also turn off colors if we're running on Windows where os.Stderr.Stat() returns an invalid handle-error
		stat, err := os.Stderr.Stat()
		if err != nil || (stat.Mode()&os.ModeCharDevice) == 0 {
			colors = false
		}
		log = srvlog.NewStdLogger(opts.Logtime, opts.Debug, opts.Trace, colors, true)
	}

	s.SetLoggerV2(log, opts.Debug, opts.Trace, opts.TraceVerbose)
}

// Returns our current logger.
func (s *Server) Logger() Logger {
	s.logging.Lock()
	defer s.logging.Unlock()
	return s.logging.logger
}

// SetLogger sets the logger of the server
func (s *Server) SetLogger(logger Logger, debugFlag, traceFlag bool) {
	s.SetLoggerV2(logger, debugFlag, traceFlag, false)
}

// SetLogger sets the logger of the server
func (s *Server) SetLoggerV2(logger Logger, debugFlag, traceFlag, sysTrace bool) {
	if debugFlag {
		atomic.StoreInt32(&s.logging.debug, 1)
	} else {
		atomic.StoreInt32(&s.logging.debug, 0)
	}
	if traceFlag {
		atomic.StoreInt32(&s.logging.trace, 1)
	} else {
		atomic.StoreInt32(&s.logging.trace, 0)
	}
	if sysTrace {
		atomic.StoreInt32(&s.logging.traceSysAcc, 1)
	} else {
		atomic.StoreInt32(&s.logging.traceSysAcc, 0)
	}
	s.logging.Lock()
	if s.logging.logger != nil {
		// Check to see if the logger implements io.Closer.  This could be a
		// logger from another process embedding the NATS server or a dummy
		// test logger that may not implement that interface.
		if l, ok := s.logging.logger.(io.Closer); ok {
			if err := l.Close(); err != nil {
				s.Errorf("Error closing logger: %v", err)
			}
		}
	}
	s.logging.logger = logger
	s.logging.Unlock()
}

// ReOpenLogFile if the logger is a file based logger, close and re-open the file.
// This allows for file rotation by 'mv'ing the file then signaling
// the process to trigger this function.
func (s *Server) ReOpenLogFile() {
	// Check to make sure this is a file logger.
	s.logging.RLock()
	ll := s.logging.logger
	s.logging.RUnlock()

	if ll == nil {
		s.Noticef("File log re-open ignored, no logger")
		return
	}

	// Snapshot server options.
	opts := s.getOpts()

	if opts.LogFile == "" {
		s.Noticef("File log re-open ignored, not a file logger")
	} else {
		fileLog := srvlog.NewFileLogger(opts.LogFile,
			opts.Logtime, opts.Debug, opts.Trace, true)
		s.SetLogger(fileLog, opts.Debug, opts.Trace)
		if opts.LogSizeLimit > 0 {
			fileLog.SetSizeLimit(opts.LogSizeLimit)
		}
		s.Noticef("File log re-opened")
	}
}

// Noticef logs a notice statement
func (s *Server) Noticef(format string, v ...interface{}) {
	s.executeLogCall(func(logger Logger, format string, v ...interface{}) {
		logger.Noticef(format, v...)
	}, format, v...)
}

// Errorf logs an error
func (s *Server) Errorf(format string, v ...interface{}) {
	s.executeLogCall(func(logger Logger, format string, v ...interface{}) {
		logger.Errorf(format, v...)
	}, format, v...)
}

// Error logs an error with a scope
func (s *Server) Errors(scope interface{}, e error) {
	s.executeLogCall(func(logger Logger, format string, v ...interface{}) {
		logger.Errorf(format, v...)
	}, "%s - %s", scope, UnpackIfErrorCtx(e))
}

// Error logs an error with a context
func (s *Server) Errorc(ctx string, e error) {
	s.executeLogCall(func(logger Logger, format string, v ...interface{}) {
		logger.Errorf(format, v...)
	}, "%s: %s", ctx, UnpackIfErrorCtx(e))
}

// Error logs an error with a scope and context
func (s *Server) Errorsc(scope interface{}, ctx string, e error) {
	s.executeLogCall(func(logger Logger, format string, v ...interface{}) {
		logger.Errorf(format, v...)
	}, "%s - %s: %s", scope, ctx, UnpackIfErrorCtx(e))
}

// Warnf logs a warning error
func (s *Server) Warnf(format string, v ...interface{}) {
	s.executeLogCall(func(logger Logger, format string, v ...interface{}) {
		logger.Warnf(format, v...)
	}, format, v...)
}

func (s *Server) RateLimitWarnf(format string, v ...interface{}) {
	statement := fmt.Sprintf(format, v...)
	if _, loaded := s.rateLimitLogging.LoadOrStore(statement, time.Now()); loaded {
		return
	}
	s.Warnf("%s", statement)
}

// Fatalf logs a fatal error
func (s *Server) Fatalf(format string, v ...interface{}) {
	s.executeLogCall(func(logger Logger, format string, v ...interface{}) {
		logger.Fatalf(format, v...)
	}, format, v...)
}

// Debugf logs a debug statement
func (s *Server) Debugf(format string, v ...interface{}) {
	if atomic.LoadInt32(&s.logging.debug) == 0 {
		return
	}

	s.executeLogCall(func(logger Logger, format string, v ...interface{}) {
		logger.Debugf(format, v...)
	}, format, v...)
}

// Tracef logs a trace statement
func (s *Server) Tracef(format string, v ...interface{}) {
	if atomic.LoadInt32(&s.logging.trace) == 0 {
		return
	}

	s.executeLogCall(func(logger Logger, format string, v ...interface{}) {
		logger.Tracef(format, v...)
	}, format, v...)
}

func (s *Server) executeLogCall(f func(logger Logger, format string, v ...interface{}), format string, args ...interface{}) {
	s.logging.RLock()
	defer s.logging.RUnlock()
	if s.logging.logger == nil {
		return
	}

	f(s.logging.logger, format, args...)
}
