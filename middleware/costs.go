package middleware

import (
	"github.com/gin-gonic/gin"
	"os"
	"time"
)

func CostHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("executionTime", time.Now().UnixMilli())
		c.Set("staticCdnPrefix", os.Getenv("STATIC_CDN_PREFIX"))
		c.Next()
	}

}
