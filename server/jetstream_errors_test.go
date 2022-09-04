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
	"errors"
	"testing"
)

func TestIsNatsErr(t *testing.T) {
	if !IsNatsErr(ApiErrors[JSNotEnabledForAccountErr], JSNotEnabledForAccountErr) {
		t.Fatalf("Expected error match")
	}

	if IsNatsErr(ApiErrors[JSNotEnabledForAccountErr], JSClusterNotActiveErr) {
		t.Fatalf("Expected error mismatch")
	}

	if IsNatsErr(ApiErrors[JSNotEnabledForAccountErr], JSClusterNotActiveErr, JSClusterNotAvailErr) {
		t.Fatalf("Expected error mismatch")
	}

	if !IsNatsErr(ApiErrors[JSNotEnabledForAccountErr], JSClusterNotActiveErr, JSNotEnabledForAccountErr) {
		t.Fatalf("Expected error match")
	}

	if !IsNatsErr(&ApiError{ErrCode: 10039}, 1, JSClusterNotActiveErr, JSNotEnabledForAccountErr) {
		t.Fatalf("Expected error match")
	}

	if IsNatsErr(&ApiError{ErrCode: 10039}, 1, 2, JSClusterNotActiveErr) {
		t.Fatalf("Expected error mismatch")
	}

	if IsNatsErr(nil, JSClusterNotActiveErr) {
		t.Fatalf("Expected error mismatch")
	}

	if IsNatsErr(errors.New("x"), JSClusterNotActiveErr) {
		t.Fatalf("Expected error mismatch")
	}
}

func TestApiError_Error(t *testing.T) {
	if es := ApiErrors[JSClusterNotActiveErr].Error(); es != "JetStream not in clustered mode (10006)" {
		t.Fatalf("Expected 'JetStream not in clustered mode (10006)', got %q", es)
	}
}

func TestApiError_NewWithTags(t *testing.T) {
	ne := NewJSRestoreSubscribeFailedError(errors.New("failed error"), "the.subject")
	if ne.Description != "JetStream unable to subscribe to restore snapshot the.subject: failed error" {
		t.Fatalf("Expected 'JetStream unable to subscribe to restore snapshot the.subject: failed error' got %q", ne.Description)
	}

	if ne == ApiErrors[JSRestoreSubscribeFailedErrF] {
		t.Fatalf("Expected a new instance")
	}
}

func TestApiError_NewWithUnless(t *testing.T) {
	if ne := NewJSStreamRestoreError(errors.New("failed error"), Unless(ApiErrors[JSNotEnabledForAccountErr])); !IsNatsErr(ne, JSNotEnabledForAccountErr) {
		t.Fatalf("Expected JSNotEnabledForAccountErr got %s", ne)
	}

	if ne := NewJSStreamRestoreError(errors.New("failed error")); !IsNatsErr(ne, JSStreamRestoreErrF) {
		t.Fatalf("Expected JSStreamRestoreErrF got %s", ne)
	}

	if ne := NewJSStreamRestoreError(errors.New("failed error"), Unless(errors.New("other error"))); !IsNatsErr(ne, JSStreamRestoreErrF) {
		t.Fatalf("Expected JSStreamRestoreErrF got %s", ne)
	}

	if ne := NewJSPeerRemapError(Unless(ApiErrors[JSNotEnabledForAccountErr])); !IsNatsErr(ne, JSNotEnabledForAccountErr) {
		t.Fatalf("Expected JSNotEnabledForAccountErr got %s", ne)
	}

	if ne := NewJSPeerRemapError(Unless(nil)); !IsNatsErr(ne, JSPeerRemapErr) {
		t.Fatalf("Expected JSPeerRemapErr got %s", ne)
	}

	if ne := NewJSPeerRemapError(Unless(errors.New("other error"))); !IsNatsErr(ne, JSPeerRemapErr) {
		t.Fatalf("Expected JSPeerRemapErr got %s", ne)
	}
}
