package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

func notPrimary(ctx *gin.Context) bool {
	if !isPrimary {
		ctx.IndentedJSON(http.StatusOK, gin.H{
			"msg": "Try Again, the controller is NOT primary currently.",
		})
		return true
	}
	return false
}
