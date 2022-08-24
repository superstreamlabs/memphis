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
)

type errOpts struct {
	err error
}

// ErrorOption configures a NATS Error helper
type ErrorOption func(*errOpts)

// Unless ensures that if err is a ApiErr that err will be returned rather than the one being created via the helper
func Unless(err error) ErrorOption {
	return func(opts *errOpts) {
		opts.err = err
	}
}

func parseOpts(opts []ErrorOption) *errOpts {
	eopts := &errOpts{}
	for _, opt := range opts {
		opt(eopts)
	}
	return eopts
}

type ErrorIdentifier uint16

// IsNatsErr determines if an error matches ID, if multiple IDs are given if the error matches any of these the function will be true
func IsNatsErr(err error, ids ...ErrorIdentifier) bool {
	if err == nil {
		return false
	}

	ce, ok := err.(*ApiError)
	if !ok || ce == nil {
		return false
	}

	for _, id := range ids {
		ae, ok := ApiErrors[id]
		if !ok || ae == nil {
			continue
		}

		if ce.ErrCode == ae.ErrCode {
			return true
		}
	}

	return false
}

// ApiError is included in all responses if there was an error.
type ApiError struct {
	Code        int    `json:"code"`
	ErrCode     uint16 `json:"err_code,omitempty"`
	Description string `json:"description,omitempty"`
}

// ErrorsData is the source data for generated errors as found in errors.json
type ErrorsData struct {
	Constant    string `json:"constant"`
	Code        int    `json:"code"`
	ErrCode     uint16 `json:"error_code"`
	Description string `json:"description"`
	Comment     string `json:"comment"`
	Help        string `json:"help"`
	URL         string `json:"url"`
	Deprecates  string `json:"deprecates"`
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("%s (%d)", e.Description, e.ErrCode)
}

func (e *ApiError) toReplacerArgs(replacements []interface{}) []string {
	var (
		ra  []string
		key string
	)

	for i, replacement := range replacements {
		if i%2 == 0 {
			key = replacement.(string)
			continue
		}

		switch v := replacement.(type) {
		case string:
			ra = append(ra, key, v)
		case error:
			ra = append(ra, key, v.Error())
		default:
			ra = append(ra, key, fmt.Sprintf("%v", v))
		}
	}

	return ra
}
