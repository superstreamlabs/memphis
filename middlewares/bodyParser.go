package middlewares

import (
	"github.com/gin-gonic/gin"
)

func Json(c *gin.Context) {
	var body interface{}
	
    if err := c.BindJSON(&body); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"message": "Request body format has to be JSON"})
	}

	c.Set("parsedBody", body)
	c.Next()
}