package models

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Schema interface {
	Validate(c *gin.Context) (validator.ValidationErrors, bool)
}