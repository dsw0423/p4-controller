package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type refresh struct {
	RefreshToken string `json:"refreshToken"`
}

func RefreshToken(c *gin.Context) {
	ref := refresh{}
	if err := c.ShouldBind(&ref); err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "refresh token miss or error: " + err.Error(),
		})
		return
	}

	h := gin.H{}
	if err := validateToken(ref.RefreshToken); err != nil {
		newRefreshToken := updateRefreshToken(ref.RefreshToken)
		if newRefreshToken != "" {
			h["refreshToken"] = newRefreshToken
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"msg": "login again"})
			return
		}
	}

	newAccessToken, err := getNewAccessToken(ref.RefreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "generate new access token error: " + err.Error()})
		return
	}
	h["accessToken"] = newAccessToken

	c.JSON(http.StatusOK, h)
}

func updateRefreshToken(token string) string {
	tk, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})

	if errors.Is(err, jwt.ErrTokenExpired) {
		claims := tk.Claims.(jwt.MapClaims)
		exp := claims["exp"].(float64)
		now := time.Now().Unix()
		if now-int64(exp) < int64(loginExpireTime) {
			claims["exp"] = time.Now().Add(expRefresh).Unix()
			tk = jwt.NewWithClaims(tk.Method, claims)
			s, err := tk.SignedString(secret)
			if err != nil {
				return ""
			}
			return s
		}
	}

	return ""
}

func getNewAccessToken(refreshToken string) (string, error) {
	username := getUsername(refreshToken)
	exp := time.Now().Add(expAccess).Unix()
	tk := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,
		"exp": exp,
	})
	return tk.SignedString(secret)
}

func getUsername(refreshToken string) string {
	tk, err := jwt.Parse(refreshToken, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil {
		return ""
	}
	claims := tk.Claims.(jwt.MapClaims)
	return claims["sub"].(string)
}
