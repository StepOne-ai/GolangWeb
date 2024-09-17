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
		if data.Username == "admin" {
			c.SetCookie(
				"adminAccess",
				"true",
				3600,
				"/",
				"localhost",
				false,
				true,
			)
		}
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
	Username string `form:"username"`
	Email string `form:"email"`
	Password string `form:"password"`
}

func RegisterPost(c *gin.Context) {
	var data FormDataReg
	c.Bind(&data)

	if data.Username == "" || data.Email == "" || data.Password == "" {
		fmt.Println("Error")
		c.Redirect(
			302,
			"/register",
		)
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
		fmt.Println("last Error")
		c.Redirect(
			302,
			"/register",
		)
	}
}

func Logout(c *gin.Context) {
	c.SetCookie("username", "", -1, "/", "localhost", false, true)
	c.Redirect(
		302,
		"/",
	)
}

func Account(c *gin.Context) {
	username := c.Param("username")

	dbPath := "./db.db"
	db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
		log.Fatal(err)
	}
		defer db.Close()

	user, err := database.GetUserByUsername(db, username)

	if err != nil {
		c.Redirect(302, "/articles")
		return
	}

	current_user, err := c.Cookie("username")

	if err != nil {
		c.Redirect(302, "/")
		return
	}

	c.HTML(
		http.StatusOK,
		"articles/account.html",
		gin.H{
			"username": user.Username,
			"email": user.Email,
			"id": user.UserID,
			"current_user": current_user,
		},
	)
}

func AccountUpdate(c *gin.Context) {
	var data FormDataReg
	c.Bind(&data)
	//fmt.Println(data.Username, data.Email, data.Password)
	current_user, err := c.Cookie("username")

	if err != nil {
		c.Redirect(302, "/")
		return
	}

	if data.Username == "" || data.Email == "" {
		fmt.Println("Error")
		c.Redirect(
			302,
			"/account/"+current_user,
		)
		return
	}

	dbPath := "./db.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	user, err := database.GetUserByUsername(db, current_user)
	if err != nil {
		log.Fatal(err)
	}

	if data.Password != "" {
		err = database.UpdateUser(db, user.UserID, data.Username, data.Email, data.Password)
		c.SetCookie(
			"username",
			data.Username,
			3600,
			"/",
			"localhost",
			false,
			true,
		)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err = database.UpdateUserWithoutPassword(db, user.UserID, data.Username, data.Email)
		c.SetCookie(
			"username",
			data.Username,
			3600,
			"/",
			"localhost",
			false,
			true,
		)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Updating possible articles written by user
	if data.Username != current_user {
		articles, err := database.GetArticlesByAuthor(db, current_user)
		if err != nil {
			log.Fatal(err)
		}

		for _, article := range articles {
			err = database.UpdateArticleAuthor(db, article.ArticleID, data.Username)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	defer db.Close()

	c.HTML(
		http.StatusOK,
		"articles/account.html",
		gin.H{
			"username": data.Username,
			"email": data.Email,
			"id": user.UserID,
			"current_user": data.Username,
		},
	)
}