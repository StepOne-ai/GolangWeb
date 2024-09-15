package database

import (
    "database/sql"
    "fmt"
    "log"

    _ "github.com/mattn/go-sqlite3"

	"golang.org/x/crypto/bcrypt"
	m "dbgolang/models"
)
func GetUsers(db *sql.DB) ([]m.User, error) {
    rows, err := db.Query("SELECT UserID, Username, Password FROM Users")
    if err != nil {
        return nil, fmt.Errorf("failed to query users: %w", err)
    }
    defer rows.Close()

    var users []m.User
    for rows.Next() {
        var u m.User
        if err := rows.Scan(&u.UserID, &u.Username, &u.PasswordHash); err != nil {
            return nil, fmt.Errorf("failed to scan row: %w", err)
        }
        users = append(users, u)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating over rows: %w", err)
    }

    return users, nil
}

func GetArticles(db *sql.DB) ([]m.Article, error) {
    rows, err := db.Query("SELECT ArticleID, Title, Author, Content FROM Articles")
    if err != nil {
        return nil, fmt.Errorf("failed to query articles: %w", err)
    }
    defer rows.Close()

    var articles []m.Article
    for rows.Next() {
        var u m.Article
        if err := rows.Scan(&u.ArticleID, &u.Title, &u.Author, &u.Content); err != nil {
            return nil, fmt.Errorf("failed to scan row: %w", err)
        }
        articles = append(articles, u)
    }

    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating over rows: %w", err)
    }

    return articles, nil
}

func CheckPassword(users []m.User, username, password string) bool {
    for _, user := range users {
        if user.Username == username && VerifyPassword(password, user.PasswordHash) {
            return true
        }
    }
    return false
}

func VerifyPassword(providedPassword, storedHash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(providedPassword))
	return err == nil
}

func CreateTableUsers(db *sql.DB) error {
    stmt, err := db.Prepare(`CREATE TABLE IF NOT EXISTS Users (
        UserID INTEGER PRIMARY KEY AUTOINCREMENT,
        Username TEXT NOT NULL,
        Email TEXT NOT NULL,
		Password TEXT NOT NULL,
        CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )`)
    if err != nil {
        return fmt.Errorf("failed to prepare create table statement: %w", err)
    }
    _, err = stmt.Exec()
    if err != nil {
        return fmt.Errorf("failed to execute create table statement: %w", err)
    }
    return nil
}

func CreateTableArticles(db *sql.DB) error {
	stmt, err := db.Prepare(`CREATE TABLE IF NOT EXISTS Articles (
		ArticleID INTEGER PRIMARY KEY AUTOINCREMENT,
		Title TEXT NOT NULL,
		Content TEXT NOT NULL,
		Author TEXT NOT NULL,
		CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("failed to prepare create table statement: %w", err)
	}
	_, err = stmt.Exec()
	if err != nil {
		return fmt.Errorf("failed to execute create table statement: %w", err)
	}
	return nil
}

func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func Login(db *sql.DB, username, password string) bool {
    users, err := GetUsers(db)

	if err != nil {	
		log.Fatalf("Failed to get users: %v", err)
	}

	if CheckPassword(users, username, password) {
		fmt.Println("Login successful")
		return true
	} else {
		fmt.Println("Login failed")
		return false
	}
}

func InsertUser(db *sql.DB, username, email string, password string) bool {
    stmt, err := db.Prepare(`INSERT INTO Users (Username, Email, Password) VALUES (?, ?, ?)`)
    if err != nil {
        return false
    }
    defer stmt.Close()

	// Hashing
	hashed, err := HashPassword(password)
	if err != nil {
		return false
	}
    _, err = stmt.Exec(username, email, hashed)
    return err != nil
}

func InsertArticle(db *sql.DB, title, content, author string) error {
	stmt, err := db.Prepare(`INSERT INTO Articles (Title, Content, Author) VALUES (?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(title, content, author)
	if err != nil {
		return fmt.Errorf("failed to execute insert statement: %w", err)
	}

	return nil
}

func GetArticleID(db *sql.DB, title string) (m.Article, error) {
    stmt, err := db.Prepare(`SELECT ArticleID FROM Articles WHERE Title = ?`)
    if err != nil {
        return m.Article{}, fmt.Errorf("failed to prepare select statement: %w", err)
    }
    defer stmt.Close()

    var article m.Article
    err = stmt.QueryRow(title).Scan(&article.ArticleID)
    if err != nil {
        return m.Article{}, fmt.Errorf("failed to execute select statement: %w", err)
    }
    fmt.Println("ArticleID: ", article.ArticleID)
    return article, nil
}

func GetArticleByID(db *sql.DB, articleID int) (m.Article, error) {
    stmt, err := db.Prepare(`SELECT ArticleID, Title, Content, Author FROM Articles WHERE ArticleID = ?`)
    if err != nil {
        return m.Article{}, fmt.Errorf("failed to prepare select statement: %w", err)
    }
    defer stmt.Close()

    var article m.Article
    err = stmt.QueryRow(articleID).Scan(&article.ArticleID, &article.Title, &article.Content, &article.Author)
    if err != nil {
        return m.Article{}, fmt.Errorf("failed to execute select statement: %w", err)
    }
    return article, nil
}

func DeleteArticle(db *sql.DB, articleID int) error {
	stmt, err := db.Prepare(`DELETE FROM Articles WHERE ArticleID = ?`)
	if err != nil {
		return fmt.Errorf("failed to prepare delete statement: %w", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(articleID)
    fmt.Println("Article deleted successfully!")
	if err != nil {
		return fmt.Errorf("failed to execute delete statement: %w", err)
	}

	return nil
}

func UpdateArticle(db *sql.DB, articleID int, title, content string) error {
    stmt, err := db.Prepare(`UPDATE Articles SET Title = ?, Content = ? WHERE ArticleID = ?`)
    if err != nil {
        return fmt.Errorf("failed to prepare update statement: %w", err)
    }
    defer stmt.Close()
    _, err = stmt.Exec(title, content, articleID)
    if err != nil {
        return fmt.Errorf("failed to execute update statement: %w", err)
    }
    fmt.Println("Article updated successfully!")
    return nil
}