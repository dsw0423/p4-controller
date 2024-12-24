package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Cors(c *gin.Context) {
	origin := c.Request.Header.Get("Origin")
	if origin != "" {
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	}

	if c.Request.Method == http.MethodOptions {
		headers := c.Request.Header.Get("Access-Control-Request-Headers")
		if headers != "" {
			c.Writer.Header().Set("Access-Control-Allow-Headers", headers)
		}

		methods := c.Request.Header.Get("Access-Control-Request-Method")
		if methods != "" {
			c.Writer.Header().Set("Access-Control-Allow-Methods", methods)
		}

		c.AbortWithStatus(http.StatusNoContent)
	}
}
