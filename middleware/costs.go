package middleware

import (
	"github.com/gin-gonic/gin"
	"time"
)

func CostHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("executionTime", time.Now().UnixMilli())
		c.Next()
	}

}
