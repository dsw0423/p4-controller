package handler

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

	tk := parts[1]
	if err := validateToken(tk); err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{
			"msg": "invalid access token: " + err.Error(),
		})
		c.Abort()
		return
	}
	log.Println("valid access token")
}

func validateToken(token string) error {
	tk, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})

	if err != nil || !tk.Valid {
		log.Println(err.Error())
	}

	return err
}
