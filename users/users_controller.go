package users

import (
	//"dbgolang/models"
	"net/http"
	//"time"
	"fmt"
	"log"
	"github.com/gin-gonic/gin"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"dbgolang/database"
)

func Login(c *gin.Context) {
	c.SetCookie("username", "", -1, "/", "localhost", false, true)
	c.HTML(
		http.StatusOK, 
		"articles/login.html", 
		nil,
	)
}

type FormData struct {
	Username string `form:"username" binding:"required"`
	Password string `form:"password" binding:"required"`
}

func LoginPost(c *gin.Context) {
	var data FormData
	c.Bind(&data)

	if data.Username == "" || data.Password == "" {
		c.Redirect(
			302,
			"/login",
		)
		return
	}

	dbPath := "./db.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if database.Login(db, data.Username, data.Password) {
		// Set cookie 
		c.SetCookie(
			"username",
			data.Username,
			3600,
			"/",
			"localhost",
			false,
			true,
		)
		c.Redirect(
			302,
			"/articles",
		)
		
		
	} else {
		// Set error
		c.HTML(
			http.StatusOK,
			"articles/login.html",
			gin.H{
				"Error": "Invalid username or password",
			},
		)
	}
}

func Register(c *gin.Context) {
	c.HTML(
		http.StatusOK,
		"articles/register.html",
		nil,
	)
}

type FormDataReg struct {
	Username string `form:"username" binding:"required"`
	Email string `form:"email" binding:"required"`
	Password string `form:"password" binding:"required"`
}

func RegisterPost(c *gin.Context) {
	var data FormDataReg
	c.Bind(&data)

	if data.Username == "" || data.Email == "" || data.Password == "" {
		c.Redirect(
			302,
			"/register",
		)
		return
	}

	dbPath := "./db.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if database.InsertUser(db, data.Username, data.Email, data.Password) {
		fmt.Println("User created successfully")
		// Set cookie
		c.SetCookie(
			"username",
			data.Username,
			3600,
			"/",
			"localhost",
			false,
			true,
		)
		c.Redirect(
			302,
			"/articles",
		)
	} else {
		// Set error
		c.Redirect(
			302,
			"/register",
		)
	}
}