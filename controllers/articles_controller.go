package controllers

import (
	"dbgolang/models"
	"net/http"
	"fmt"
	"log"
	"github.com/gin-gonic/gin"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"dbgolang/database"
	"strconv"
)

func ArticlesIndex(c *gin.Context) {
	username, err := c.Cookie("username")
	if username == "" || err != nil {
		c.Redirect(302, "/")
		return
	}

	dbPath := "./db.db"
    db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	var articles2 []models.Article
	articles2, err = database.GetArticles(db)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	articles := []models.Article{}

	if len(articles2) != 0 {
		articles = append(articles, articles2...)
	}

	avatar_url, _ := database.GetAvatarURLByUsername(db, username)

	c.HTML(
		http.StatusOK,
		"articles/index.html",
		gin.H{
			"avatar_url": avatar_url,
			"articles": articles,
			"username": username,
		},
	)
}

type FormData struct {
	Title   string `form:"title"`
	Content string `form:"content"`
}

func ArticlesCreate(c *gin.Context) {
	var data FormData
	c.Bind(&data)
	fmt.Println("Title: ", data.Title)
	// * Insert article
	dbPath := "./db.db"
    db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	username, err := c.Cookie("username")
	if err != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}
	err = database.InsertArticle(db, data.Title, data.Content, username)
	if err != nil {
		log.Fatal(err)
	}
	article, err := database.GetArticleID(db, data.Title)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	c.HTML(http.StatusOK,
		"articles/article.html",
		gin.H {
			"Title": data.Title,
			"Content": data.Content,
			"Author": username,
			"ArticleID": article.ArticleID,
		})
}

func ArticleDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Redirect(500, "/")
	}

	// Get article from database
	dbPath := "./db.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		c.Redirect(500, "/")
	}

	err = database.DeleteArticle(db, id)
	// Delete article from database
	if err != nil {
		c.Redirect(500, "/")
	}
	// Close database
	defer db.Close()
	
	// Redirect to articles index
	c.Redirect(302, "/articles")
}

func ArticleUpdate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.Redirect(500, "/")
	}
	// Get article from database
	dbPath := "./db.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		c.Redirect(500, "/")
	}

	article, err := database.GetArticleByID(db, id)
	// Delete article from database
	if err != nil {
		c.Redirect(500, "/")
	}
	// Close database
	defer db.Close()
	// Redirect to articles index
	c.HTML(http.StatusOK,
		"articles/update.html",
		gin.H {
			"Title": article.Title,
			"Content": article.Content,
			"ArticleID": article.ArticleID,
		})
}

type FormDataWithID struct {
	ID      int    `form:"ID" binding:"required"`
	Title   string `form:"title" binding:"required"`
	Content string `form:"content" binding:"required"`
}

func ArticleUpdatePost(c *gin.Context) {
	if c.Request.Method != "POST" {
		c.Redirect(302, "/")
	}
	_, err := c.Cookie("username") 
	if err != nil {
		c.Redirect(302, "/")
	}
	dbPath := "./db.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		c.Redirect(500, "/")
	}
	var data FormDataWithID
	c.Bind(&data)
	if data.Title == "" || data.Content == "" {
		c.HTML(400, "articles/error.html", gin.H{
			"error": "Title and content are required",
		})
		return
	}
	fmt.Println("Title: ", data.Title)
	fmt.Println("Content: ", data.Content)
	fmt.Println("ID: ", data.ID)
	err = database.UpdateArticle(db, data.ID, data.Title, data.Content)
	if err != nil {
		c.Redirect(500, "/")
	}
	// Close database
	defer db.Close()
	// Redirect to articles index
	c.Redirect(302, "/articles")
}

func ArticleShow(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.Redirect(500, "/")
    }
    // Get article from database
    dbPath := "./db.db"
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        c.Redirect(500, "/")
    }
    article, err := database.GetArticleByID(db, id)
    // Delete article from database
    if err != nil {
        c.Redirect(500, "/")
    }
	//Get username
	username, err := c.Cookie("username")
	if err != nil {
		c.Redirect(500, "/")
		return
	}
	avatar_url, _ := database.GetAvatarURLByUsername(db, username)
    // Close database
    defer db.Close()
    c.HTML(200, "articles/article_show.html", gin.H{
		"ArticleID": article.ArticleID,
        "Title":   article.Title,
        "Content": article.Content,
        "Author":  article.Author,
		"username": username,
		"avatar_url": avatar_url,
    })
}