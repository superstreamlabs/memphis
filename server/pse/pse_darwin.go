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
package pse

// On macs after some studying it seems that typical tools like ps and activity monitor report MaxRss and not
// current RSS. I wrote some C code to pull the real RSS and although it does not go down very often, when it does
// that is not reflected in the typical tooling one might compare us to, so we can skip cgo and just use rusage imo.
// We also do not use virtual memory in the upper layers at all, so ok to skip since rusage does not report vss.

import (
	"math"
	"sync"
	"syscall"
	"time"
)

type lastUsage struct {
	sync.Mutex
	last time.Time
	cpu  time.Duration
	rss  int64
	pcpu float64
}

// To hold the last usage and call time.
var lu lastUsage

func init() {
	updateUsage()
	periodic()
}

// Get our usage.
func getUsage() (now time.Time, cpu time.Duration, rss int64) {
	var ru syscall.Rusage
	syscall.Getrusage(syscall.RUSAGE_SELF, &ru)
	now = time.Now()
	cpu = time.Duration(ru.Utime.Sec)*time.Second + time.Duration(ru.Utime.Usec)*time.Microsecond
	cpu += time.Duration(ru.Stime.Sec)*time.Second + time.Duration(ru.Stime.Usec)*time.Microsecond
	return now, cpu, ru.Maxrss
}

// Update last usage.
// We need to have a prior sample to compute pcpu.
func updateUsage() (pcpu float64, rss int64) {
	lu.Lock()
	defer lu.Unlock()

	now, cpu, rss := getUsage()
	// Don't skew pcpu by sampling too close to last sample.
	if elapsed := now.Sub(lu.last); elapsed < 500*time.Millisecond {
		// Always update rss.
		lu.rss = rss
	} else {
		tcpu := float64(cpu - lu.cpu)
		lu.last, lu.cpu, lu.rss = now, cpu, rss
		// Want to make this one decimal place and not count on upper layers.
		// Cores already taken into account via cpu time measurements.
		lu.pcpu = math.Round(tcpu/float64(elapsed)*1000) / 10
	}
	return lu.pcpu, lu.rss
}

// Sampling function to keep pcpu relevant.
func periodic() {
	updateUsage()
	time.AfterFunc(time.Second, periodic)
}

// ProcUsage returns CPU and memory usage.
// Note upper layers do not use virtual memory size, so ok that it is not filled in here.
func ProcUsage(pcpu *float64, rss, vss *int64) error {
	*pcpu, *rss = updateUsage()
	return nil
}
