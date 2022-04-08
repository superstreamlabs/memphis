package utils

import (
	"fmt"
	"memphis-control-plane/logger"
	"mime/multipart"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
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
			logger.Error("validateSchema error: " + err.Error())
			c.AbortWithStatusJSON(400, gin.H{"message": "Could not complete uploading your file, please check your file"})
			return nil, true
		}

		fileExt := filepath.Ext(uploadedFile.Filename)
		if fileExt != ".png" && fileExt != ".jpg" && fileExt != ".jpeg" {
			c.AbortWithStatusJSON(400, gin.H{"message": "You can upload only png,jpg or jpeg file formats"})
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
