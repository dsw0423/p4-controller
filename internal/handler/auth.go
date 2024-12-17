package handler

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func AuthCheck(c *gin.Context) {
	auth := c.GetHeader("Authorization")
	parts := strings.Split(auth, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		log.Println("Auth:", auth)
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "token error",
		})
		c.Abort()
		return
	}

	s := parts[1]
	tk, err := jwt.Parse(s, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})

	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusForbidden, gin.H{
			"msg": "token parse error: " + err.Error(),
		})
		c.Abort()
		return
	}

	if claims, ok := tk.Claims.(jwt.MapClaims); ok {
		fmt.Println("User", claims["username"], "authorized")
	}
}
