package handler

import (
	"database/sql"
	"encoding/base64"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

var (
	secret = []byte("dsw")
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

	var encoded string
	query := `select password from p4ctl_users where username = ?`
	err = db.QueryRow(query, user.Username).Scan(&encoded)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "execute sql error: " + err.Error(),
		})
		return
	}

	if err := compareEncodedAndPassword(encoded, user.Password); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"msg": "user name or password error: " + err.Error(),
		})
		return
	}

	tk := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": user.Username,
		})
	s, err := tk.SignedString(secret)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": err.Error(),
		})
		return
	}

	log.Println("JWT:", s)
	c.JSON(http.StatusOK, gin.H{
		"token": s,
	})
}

func compareEncodedAndPassword(encodedBase64, password string) error {
	encoded, err := base64.StdEncoding.DecodeString(encodedBase64)
	if err != nil {
		return err
	}

	return bcrypt.CompareHashAndPassword(encoded, []byte(password))
}
