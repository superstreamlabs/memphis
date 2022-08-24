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
	"strings"
	"testing"
)

func TestErrCtx(t *testing.T) {
	ctx := "Extra context information"
	e := NewErrorCtx(ErrWrongGateway, ctx)

	if e.Error() != ErrWrongGateway.Error() {
		t.Fatalf("%v and %v are supposed to be identical", e, ErrWrongGateway)
	}
	if e == ErrWrongGateway {
		t.Fatalf("%v and %v can't be compared this way", e, ErrWrongGateway)
	}
	if !ErrorIs(e, ErrWrongGateway) {
		t.Fatalf("%s and %s ", e, ErrWrongGateway)
	}
	if UnpackIfErrorCtx(ErrWrongGateway) != ErrWrongGateway.Error() {
		t.Fatalf("Error of different type should be processed unchanged")
	}
	trace := UnpackIfErrorCtx(e)
	if !strings.HasPrefix(trace, ErrWrongGateway.Error()) {
		t.Fatalf("original error needs to remain")
	}
	if !strings.HasSuffix(trace, ctx) {
		t.Fatalf("ctx needs to be added")
	}
}

func TestErrCtxWrapped(t *testing.T) {
	ctxO := "Original Ctx"
	eO := NewErrorCtx(ErrWrongGateway, ctxO)
	ctx := "Extra context information"
	e := NewErrorCtx(eO, ctx)

	if e.Error() != ErrWrongGateway.Error() {
		t.Fatalf("%v and %v are supposed to be identical", e, ErrWrongGateway)
	}
	if e == ErrWrongGateway {
		t.Fatalf("%v and %v can't be compared this way", e, ErrWrongGateway)
	}
	if !ErrorIs(e, ErrWrongGateway) {
		t.Fatalf("%s and %s ", e, ErrWrongGateway)
	}
	if UnpackIfErrorCtx(ErrWrongGateway) != ErrWrongGateway.Error() {
		t.Fatalf("Error of different type should be processed unchanged")
	}
	trace := UnpackIfErrorCtx(e)
	if !strings.HasPrefix(trace, ErrWrongGateway.Error()) {
		t.Fatalf("original error needs to remain")
	}
	if !strings.HasSuffix(trace, ctx) {
		t.Fatalf("ctx needs to be added")
	}
	if !strings.Contains(trace, ctxO) {
		t.Fatalf("Needs to contain every context")
	}
}
