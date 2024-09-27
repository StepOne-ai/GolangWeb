package main

import (
	"database/sql"
	"dbgolang/betting"
	"dbgolang/controllers"
	"dbgolang/database"
	u "dbgolang/users"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
    dbPath := "./db.db"
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        log.Fatalf("Failed to open database: %v", err)
    }
    defer db.Close()

    var count int64 = 86400

    // Create tables
    err = database.CreateTableUsers(db)
    if err != nil {
        log.Fatalf("Failed to create table: %v", err)
    }
    err = database.CreateTableArticles(db)
    if err != nil {
        log.Fatalf("Failed to create table: %v", err)
    }
    err = database.CreateTableCandidates(db)
    if err != nil {
        log.Fatalf("Failed to create table: %v", err)
    }
    err = database.CreateTableVotes(db)
    if err != nil {
        log.Fatalf("Failed to create table: %v", err)
    }
    err = database.CreateTableWallets(db)
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
    r.POST("/betting/new", betting.BettingPost)

    r.POST("/vote/win/:id", betting.VoteWin)
    r.POST("/vote/lose/:id", betting.VoteLose)
    r.GET("/vote/clear/:id", betting.VoteClear)

    //Handle user logout
    r.GET("/logout", u.Logout)

    // // Set Admin access
    // r.GET("/saa", func(c *gin.Context) {
    //     username, err := c.Cookie("username")
    //     if err != nil {
    //         c.Redirect(302, "/")
    //     }
    //     if username == "admin" {
    //         c.SetCookie("adminAccess", "true", 3600, "/", "localhost", false, true)
    //         c.Redirect(302, "/betting")
    //     } else {
    //         c.Redirect(302, "/")
    //     }
    // })

    // Remove Admin access
    r.GET("/raa", func(c *gin.Context) {
        c.SetCookie("adminAccess", "false", -1, "/", "localhost", false, true)
        c.Redirect(302, "/betting")
    })

    var time2 int64 = 1727611263
    r.GET("/timer", func(ctx *gin.Context) {
        currentTime := time.Now()
        seconds := currentTime.Unix()
        time := time2 + count - seconds
        day := strconv.Itoa(int(time/(60*60*24)))
        hours := strconv.Itoa(int(time/(60*60)%24))
        minutes := strconv.Itoa(int(time/(60)%60))
        secs := strconv.Itoa(int(time%60))
        ctx.HTML(200, "articles/timer.html", gin.H{
            "time": day + "d " + hours + "h " + minutes + "m " + secs + "s",
        })
    })

    r.GET("/results", func(ctx *gin.Context) {
        if time2 - count + time.Now().Unix() == 0 {
            // 1. Simulate results 50/50 for each candidate
            // 2. Remove all candidates
            // 3. Pay out to users based on coefficients
            return
        } else {
            ctx.Redirect(302, "/")
            return
        }
    })

	log.Println("Server started at localhost:3000")
	r.Run(":3000")
}
