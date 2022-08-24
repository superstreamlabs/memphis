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
package test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestPidFile(t *testing.T) {
	opts := DefaultTestOptions

	tmpDir := createDir(t, "_nats-server")
	defer removeDir(t, tmpDir)

	file := createFileAtDir(t, tmpDir, "nats-server:pid_")
	file.Close()
	opts.PidFile = file.Name()

	s := RunServer(&opts)
	s.Shutdown()

	buf, err := ioutil.ReadFile(opts.PidFile)
	if err != nil {
		t.Fatalf("Could not read pid_file: %v", err)
	}
	if len(buf) <= 0 {
		t.Fatal("Expected a non-zero length pid_file")
	}

	pid := 0
	fmt.Sscanf(string(buf), "%d", &pid)
	if pid != os.Getpid() {
		t.Fatalf("Expected pid to be %d, got %d\n", os.Getpid(), pid)
	}
}
