package main

import (
    "database/sql"
    "fmt"
    "log"
    "dbgolang/controllers"
    _ "github.com/mattn/go-sqlite3"
    "dbgolang/database"
    "github.com/gin-gonic/gin"
    u "dbgolang/users"
    "dbgolang/betting"
)

func main() {
    dbPath := "./db.db"
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        log.Fatalf("Failed to open database: %v", err)
    }
    defer db.Close()

    // Create tables
    err = database.CreateTableUsers(db)
    if err != nil {
        log.Fatalf("Failed to create table: %v", err)
    }
    err = database.CreateTableArticles(db)
    if err != nil {
        log.Fatalf("Failed to create table: %v", err)
    }

    fmt.Println("Tables created successfully")


    // * Server side 
    r := gin.Default()
	r.Use(gin.Logger())

	r.LoadHTMLGlob("views/**/*")
	
	r.GET("/", u.Login)
    r.GET("/login", func (c *gin.Context) {
        c.SetCookie("username", "", -1, "/", "localhost", false, true)
        c.Redirect(302, "/")
    })
    r.GET("/articles", controllers.ArticlesIndex)

	r.POST("/articles/new", controllers.ArticlesCreate)
    r.POST("/login/new", u.LoginPost)

    r.GET("/register", u.Register)
    r.POST("/register/new", u.RegisterPost)

    // Handle article delete
    r.GET("/articles/delete/:id", controllers.ArticleDelete)
    // Handle article update
    r.GET("/articles/update/:id", controllers.ArticleUpdate)
    r.POST("/articles/update/new", controllers.ArticleUpdatePost)
    // Handle Article show
    r.GET("/articles/show/:id", controllers.ArticleShow)

    //Handle user account
    r.GET("/account/:username", u.Account)
    r.POST("/account/update/new", u.AccountUpdate)

    //Handle betting system
    r.GET("/betting", betting.BettingIndex)

    //Handle user logout
    r.GET("/logout", u.Logout)

	log.Println("Server started at localhost:8080")
	r.Run(":3000")
    // Insert user
    //username := "john_doe"
    // email := "john@example.com"
	// password := "password123"

    // err = insertUser(db, username, email, password)
    // if err != nil {
    //     log.Fatalf("Failed to insert user: %v", err)
    // }

    // fmt.Println("User inserted successfully")

}
