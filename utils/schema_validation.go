// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.
package utils

import (
	"fmt"

	"mime/multipart"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

const (
	SHOWABLE_ERROR_STATUS_CODE = 406
)

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
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "Could not complete uploading your file, please check your file"})
			return nil, true
		}

		fileExt := filepath.Ext(uploadedFile.Filename)
		if fileExt != ".png" && fileExt != ".jpg" && fileExt != ".jpeg" {
			// logger.Warn("You can upload only png,jpg or jpeg file formats")
			c.AbortWithStatusJSON(SHOWABLE_ERROR_STATUS_CODE, gin.H{"message": "You can upload only png,jpg or jpeg file formats"})
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
