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

//go:build windows
// +build windows

package pse

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func checkValues(t *testing.T, pcpu, tPcpu float64, rss, tRss int64) {
	if pcpu != tPcpu {
		delta := int64(pcpu - tPcpu)
		if delta < 0 {
			delta = -delta
		}
		if delta > 30 { // 30%?
			t.Fatalf("CPUs did not match close enough: %f vs %f", pcpu, tPcpu)
		}
	}
	if rss != tRss {
		delta := rss - tRss
		if delta < 0 {
			delta = -delta
		}
		if delta > 1024*1024 { // 1MB
			t.Fatalf("RSSs did not match close enough: %d vs %d", rss, tRss)
		}
	}
}

func TestPSEmulationWin(t *testing.T) {
	var pcpu, tPcpu float64
	var rss, vss, tRss int64

	runtime.GC()

	if err := ProcUsage(&pcpu, &rss, &vss); err != nil {
		t.Fatalf("Error:  %v", err)
	}

	runtime.GC()

	imageName := getProcessImageName()
	// query the counters using typeperf
	out, err := exec.Command("typeperf.exe",
		fmt.Sprintf("\\Process(%s)\\%% Processor Time", imageName),
		fmt.Sprintf("\\Process(%s)\\Working Set - Private", imageName),
		fmt.Sprintf("\\Process(%s)\\Virtual Bytes", imageName),
		"-sc", "1").Output()
	if err != nil {
		t.Fatal("unable to run command", err)
	}

	// parse out results - refer to comments in procUsage for detail
	results := strings.Split(string(out), "\r\n")
	values := strings.Split(results[2], ",")

	// parse pcpu
	tPcpu, err = strconv.ParseFloat(strings.Trim(values[1], "\""), 64)
	if err != nil {
		t.Fatalf("Unable to parse percent cpu: %s", values[1])
	}

	// parse private bytes (rss)
	fval, err := strconv.ParseFloat(strings.Trim(values[2], "\""), 64)
	if err != nil {
		t.Fatalf("Unable to parse private bytes: %s", values[2])
	}
	tRss = int64(fval)

	checkValues(t, pcpu, tPcpu, rss, tRss)

	runtime.GC()

	// Again to test caching
	if err = ProcUsage(&pcpu, &rss, &vss); err != nil {
		t.Fatalf("Error:  %v", err)
	}
	checkValues(t, pcpu, tPcpu, rss, tRss)
}
