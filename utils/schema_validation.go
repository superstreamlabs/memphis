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
package utils

import (
	"fmt"

	"memphis-broker/conf"

	"mime/multipart"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var configuration = conf.GetConfig()

type ValidationError struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

func descriptive(verr validator.ValidationErrors) []ValidationError {
	errs := []ValidationError{}

	for _, f := range verr {
		err := f.ActualTag()
		if f.Param() != "" {
			err = fmt.Sprintf("%s=%s", err, f.Param())
		}
		errs = append(errs, ValidationError{Field: f.Field(), Reason: err})
	}

	return errs
}

func InitializeValidations() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
	}
}

func validateSchema(c *gin.Context, schema interface{}, containFile bool, file *multipart.FileHeader) (validator.ValidationErrors, bool) {
	if c.Request.Method == "GET" {
		if err := c.ShouldBind(schema); err != nil {
			if verr, ok := err.(validator.ValidationErrors); ok {
				return verr, true
			}
		}
	} else if containFile {
		uploadedFile, err := c.FormFile("file")
		if err != nil {
			// logger.Error("validateSchema error: " + err.Error())
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Could not complete uploading your file, please check your file"})
			return nil, true
		}

		fileExt := filepath.Ext(uploadedFile.Filename)
		if fileExt != ".png" && fileExt != ".jpg" && fileExt != ".jpeg" {
			// logger.Warn("You can upload only png,jpg or jpeg file formats")
			c.AbortWithStatusJSON(configuration.SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "You can upload only png,jpg or jpeg file formats"})
			return nil, true
		}

		*file = *uploadedFile
		return nil, false
	} else if err := c.ShouldBindJSON(schema); err != nil {
		if verr, ok := err.(validator.ValidationErrors); ok {
			return verr, true
		}

		c.AbortWithStatusJSON(400, gin.H{"message": "Body params have to be in JSON format"})
		return nil, true
	}

	return nil, false
}

func Validate(c *gin.Context, schema interface{}, containFile bool, file *multipart.FileHeader) bool {
	verr, errorExist := validateSchema(c, schema, containFile, file)
	if verr != nil {
		c.AbortWithStatusJSON(400, gin.H{"message": descriptive(verr)})
		return false
	}

	return !errorExist
}
