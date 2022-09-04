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

//go:build openbsd
// +build openbsd

package server

import (
	"os"
	"syscall"
)

func diskAvailable(storeDir string) int64 {
	var ba int64
	if _, err := os.Stat(storeDir); os.IsNotExist(err) {
		os.MkdirAll(storeDir, defaultDirPerms)
	}
	var fs syscall.Statfs_t
	if err := syscall.Statfs(storeDir, &fs); err == nil {
		// Estimate 75% of available storage.
		ba = int64(uint64(fs.F_bavail) * uint64(fs.F_bsize) / 4 * 3)
	} else {
		// Used 1TB default as a guess if all else fails.
		ba = JetStreamMaxStoreDefault
	}
	return ba
}
