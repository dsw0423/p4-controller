package main

import (
	"net/http"
	"strconv"
	"strings"

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

func stringToByteSlice(s string) []byte {
	ss := strings.Split(s, ":")
	data := make([]byte, len(ss))
	for i, b := range ss {
		d, _ := strconv.ParseUint(b, 16, 8)
		data[i] = byte(d)
	}
	return data
}
