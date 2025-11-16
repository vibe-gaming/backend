package apiHttp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func corsMiddleware(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "*")
	c.Header("Access-Control-Allow-Headers", "*")

	if c.Request.Method != http.MethodOptions {
		c.Next()
	} else {
		c.AbortWithStatus(http.StatusOK)
	}
}
