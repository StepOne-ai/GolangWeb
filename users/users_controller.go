package users

import (
	//"dbgolang/models"
	// "bytes"

	"io"
	// "mime/multipart"

	"net/http"
	"os"
	"regexp"

	//"time"
	"database/sql"
	"dbgolang/database"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
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
		fmt.Println("Login successful: ", data.Username)
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
		return

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
	Email    string `form:"email"`
	Password string `form:"password"`
}

func isStrongPassword(password string) bool {
	// Define the requirements for a strong password
	minLength := 8

	// Use regular expressions to check the password
	uppercaseRegex := regexp.MustCompile(`[A-Z]`)
	lowercaseRegex := regexp.MustCompile(`[a-z]`)
	numberRegex := regexp.MustCompile(`[0-9]`)
	specialCharRegex := regexp.MustCompile(`[^A-Za-z0-9]`)

	// Check the length of the password
	if len(password) < minLength {
		return false
	}

	// Check for uppercase letters
	if !uppercaseRegex.MatchString(password) {
		return false
	}

	// Check for lowercase letters
	if !lowercaseRegex.MatchString(password) {
		return false
	}

	// Check for numbers
	if !numberRegex.MatchString(password) {
		return false
	}

	// Check for special characters
	if !specialCharRegex.MatchString(password) {
		return false
	}

	// If all checks pass, the password is strong
	return true
}

func RegisterPost(c *gin.Context) {
	var data FormDataReg
	c.Bind(&data)

	// Check password min length

	if !isStrongPassword(data.Password) {
		c.HTML(
			http.StatusOK,
			"articles/error.html",
			gin.H{
				"error": "Пароль недостаточно сильный!",
			},
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
		return
	} else {
		fmt.Println("last Error")
		c.Redirect(
			302,
			"/register",
		)
		return
	}
}

func Logout(c *gin.Context) {
	c.SetCookie("username", "", -1, "/", "localhost", false, true)
	c.SetCookie("adminAccess", "", -1, "/", "localhost", false, true)
	c.Redirect(
		302,
		"/",
	)
}

func Account(c *gin.Context) {
	username := c.Param("username")

	current_user, err := c.Cookie("username")

	if err != nil {
		c.Redirect(302, "/")
		return
	}

	if current_user == "" {
		c.Redirect(302, "/articles")
		return
	}

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

	balance, err := database.GetBalanceByUserID(db, user.UserID)

	if err != nil {
		log.Fatal(err)
	}

	avatar_url, err := database.GetAvatarByUsername(db, user.Username)
	if err != nil {
		avatar_url = "https://yandex.ru/images/search?pos=1&from=tabbar&img_url=https%3A%2F%2Fyt3.googleusercontent.com%2Fytc%2FAIdro_k8ktKuQmVRXjH3RzMekX2wCP6VoKl3qiVYk7TZGmTl850%3Ds900-c-k-c0x00ffffff-no-rj&text=default+avatar&rpt=simage&lr=160857"
	}

	c.HTML(
		http.StatusOK,
		"articles/account.html",
		gin.H{
			"username":     user.Username,
			"email":        user.Email,
			"id":           user.UserID,
			"balance":      balance,
			"current_user": current_user,
			"avatar_url":   avatar_url,
		},
	)
}

type FormDataAccount struct {
	Username     string `form:"username"`
	Email        string `form:"email"`
	Password     string `form:"password"`
	Old_password string `form:"old_password"`
	TopUp        int    `form:"balance"`
}

func AccountUpdate(c *gin.Context) {
	var data FormDataAccount
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

	if data.Password != "" && data.Old_password != "" {
		// Check password
		user_hash, _ := database.GetUserByUsername(db, current_user)
		if !database.VerifyPassword(data.Old_password, user_hash.PasswordHash) {
			c.HTML(
				http.StatusBadRequest,
				"articles/error.html",
				gin.H{
					"error": "Неверный пароль!",
				},
			)
			return
		}
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
	// file, err := c.FormFile("avatar")
	// fmt.Println(file.Filename)
	// if err != nil {
	// 	c.HTML(
	// 		http.StatusOK,
	// 		"articles/error.html",
	// 		gin.H{
	// 			"error": "Не удалось загрузить аватар!",
	// 		})
	// 	return
	// }
	// const (
	// 	CLIENT_ID = "X3GY6WP6R5JAFH2LRwel"
	// 	SECRET_KEY = "gDt1N6Y3abo59arPoKNVPSnWH9f5RqL1kaa"
	// )
	// if file != nil {
	// 	req, err := http.NewRequest("POST", "https://api.imageban.ru/v1", nil)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	authToken := "Bearer " + CLIENT_ID
	// 	req.Header.Set("Authorization", authToken)

	// 	formData := new(bytes.Buffer)
	// 	writer := multipart.NewWriter(formData)
	// 	part, err := writer.CreateFormFile("file", file.Filename)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	file, _ := file.Open()

	// 	_, err = io.Copy(part, file)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	err = writer.Close()
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	req.Body = io.NopCloser(formData)

	// 	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 	client := &http.Client{}
	// 	resp, err := client.Do(req)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	type ImageData struct {
	// 		ID string `json:"id"`
	// 		Date string `json:"date"`
	// 		Name string `json:"name"`
	// 		Server string `json:"server"`
	// 		Views string `json:"views"`
	// 		Description string `json:"description"`
	// 		ImgName string `json:"img_name"`
	// 		Favorite bool `json:"favorite"`
	// 		Size string `json:"size"`
	// 		Resolution string `json:"resolution"`
	// 		Link string `json:"link"`
	// 		ShortLink string `json:"short_link"`
	// 	}

	// 	type Response struct {
	// 		Data []ImageData `json:"data"`
	// 		Success bool `json:"success"`
	// 		Status int `json:"status"`
	// 	}

	// 	if resp.StatusCode == 200 {
	// 		var response Response
	// 		err = json.NewDecoder(resp.Body).Decode(&response)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		responseData := response.Data[0].Link
	// 		fmt.Println(responseData)
	// 	}
	// }
	// Updating balance
	if data.TopUp >= 0 {
		err = database.UpdateBalance(db, user.UserID, data.TopUp)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		c.HTML(
			http.StatusOK,
			"articles/error.html",
			gin.H{
				"error": "Сумма пополнения должна быть положительной!",
			},
		)
	}

	balance, err := database.GetBalanceByUserID(db, user.UserID)
	if err != nil {
		log.Fatal(err)
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

	avatar_url, err := database.GetAvatarByUsername(db, data.Username)
	if err != nil {
		avatar_url = "https://yandex.ru/images/search?pos=1&from=tabbar&img_url=https%3A%2F%2Fyt3.googleusercontent.com%2Fytc%2FAIdro_k8ktKuQmVRXjH3RzMekX2wCP6VoKl3qiVYk7TZGmTl850%3Ds900-c-k-c0x00ffffff-no-rj&text=default+avatar&rpt=simage&lr=160857"
	}

	c.HTML(
		http.StatusOK,
		"articles/account.html",
		gin.H{
			"username":     data.Username,
			"email":        data.Email,
			"id":           user.UserID,
			"balance":      balance,
			"current_user": data.Username,
			"avatar_url":   avatar_url,
		},
	)
}

func AccountAvatar(c *gin.Context) {
	current_user, err := c.Cookie("username")
	if err != nil {
		c.Redirect(302, "/")
	}

	avatar, _ := c.FormFile("avatar")
	if avatar != nil {
		fmt.Println(avatar.Filename)
		dbPath := "./db.db"
		db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		user, _ := database.GetUserByUsername(db, current_user)

		tmpfile, err := os.Create("./static/images/" + current_user + ".jpeg")
		defer tmpfile.Close()
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		file, err := avatar.Open()
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = io.Copy(tmpfile, file)
		if err != nil {
			c.Writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		balance, err := database.GetBalanceByUserID(db, user.UserID)
		if err != nil {
			log.Fatal(err)
		}
		// Avatar to db
		avatar_url, err := database.GetAvatarByUsername(db, current_user)
		if avatar_url != "" {
			err = database.UpdateAvatar(db, user.UserID, "/static/images/"+current_user+".jpeg")
			if err != nil {
				log.Fatal(err)
			}
			c.HTML(http.StatusOK, "articles/account.html", gin.H{
				"username":     current_user,
				"email":        user.Email,
				"id":           user.UserID,
				"balance":      balance,
				"current_user": current_user,
				"avatar_url":   "/static/images/" + current_user + ".jpeg",
			})
			return
		}
		err = database.CreateAvatar(db, user.UserID, "/static/images/"+current_user+".jpeg")
		if err != nil {
			log.Fatal(err)
		}
		c.HTML(http.StatusOK, "articles/account.html", gin.H{
			"username":     current_user,
			"email":        user.Email,
			"id":           user.UserID,
			"balance":      balance,
			"current_user": current_user,
			"avatar_url":   "/static/images/" + current_user + ".jpeg",
		})
	} else {
		c.Redirect(302, "/account/"+current_user)
	}
}
