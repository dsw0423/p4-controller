package handler

import (
	"database/sql"
	"encoding/base64"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	secret          = []byte("dsw")
	expAccess       = 5 * time.Minute
	expRefresh      = 7 * time.Hour * 24
	loginExpireTime = 3 * 24 * 60 * 60 // 3 day
)

type userLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(c *gin.Context) {
	db, err := sql.Open("mysql", "root:0423@(127.0.0.1:3306)/test?parseTime=true")
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "cannot connect to database: " + err.Error(),
		})
		return
	}
	defer db.Close()

	user := userLogin{}
	if err := c.ShouldBind(&user); err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": err.Error(),
		})
		return
	}
	log.Println(user.Username + " : " + user.Password)

	/* TODO: password should be encrypted in transport*/
	var encoded string
	query := `select password from p4ctl_users where username = ?`
	err = db.QueryRow(query, user.Username).Scan(&encoded)
	if err != nil {
		log.Println(err.Error())
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusForbidden, gin.H{"msg": "username or password error: " + err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"msg": "execute sql error: " + err.Error(),
			})
		}
		return
	}

	if err := compareEncodedAndPassword(encoded, user.Password); err != nil {
		c.JSON(http.StatusForbidden, gin.H{
			"msg": "username or password error: " + err.Error(),
		})
		return
	}

	accessToken, refreshToken, err := generateTokens(user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

func compareEncodedAndPassword(encodedBase64, password string) error {
	encrypted, err := base64.StdEncoding.DecodeString(encodedBase64)
	if err != nil {
		return err
	}

	return bcrypt.CompareHashAndPassword(encrypted, []byte(password))
}

func generateTokens(username string) (string, string, error) {
	exp := time.Now().Add(expAccess).Unix()
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,
		"exp": exp,
	})
	accessTokenString, err := accessToken.SignedString(secret)
	if err != nil {
		return "", "", err
	}

	exp = time.Now().Add(expRefresh).Unix()
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,
		"exp": exp,
	})
	refreshTokenString, err := refreshToken.SignedString(secret)
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}
