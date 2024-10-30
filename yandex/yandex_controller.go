package yandex

import (
	"bytes"
	"database/sql"
	"dbgolang/database"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	m "dbgolang/models"

	"github.com/gin-gonic/gin"
)

const (
	clientID     = "05cfb5d9fc7e480c9cd278a8f1fbf5e8"
	clientSecret = "905bdc0de6d04031b0b309c3344c86ef"
	redirectURI  = "http://localhost/callback_yandex"
	authURL      = "https://oauth.yandex.com/authorize"
	tokenURL     = "https://oauth.yandex.ru/token"
)

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	TokenType   string `json:"token_type"`
}

type UserResponse struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

// Config holds the configuration for the Yandex module
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

func YandexLogin(c *gin.Context) {
	yand, err := New(Config{ClientID: clientID, ClientSecret: clientSecret, RedirectURI: redirectURI})
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	yand.LoginHandler(c)
}

func YandexCallback(c *gin.Context) {
	yand, err := New(Config{ClientID: clientID, ClientSecret: clientSecret, RedirectURI: redirectURI})
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	yand.CallbackHandler(c)
}

// New returns a new Yandex module instance
func New(config Config) (*Yandex, error) {
	if config.ClientID == "" || config.ClientSecret == "" || config.RedirectURI == "" {
		return nil, fmt.Errorf("missing configuration")
	}
	return &Yandex{config: config}, nil
}

// Yandex is the main struct for the Yandex module
type Yandex struct {
	config Config
}

// LoginHandler returns an HTTP handler for the login page
func (y *Yandex) LoginHandler(c *gin.Context) {
	authURLParams := url.Values{
		"client_id":     {y.config.ClientID},
		"redirect_uri":  {y.config.RedirectURI},
		"response_type": {"code"},
	}
	authURLWithParams := authURL + "?" + authURLParams.Encode()
	http.Redirect(c.Writer, c.Request, authURLWithParams, http.StatusFound)
}

type User struct {
	ID              string `json:"id"`
	Login           string `json:"login"`
	ClientID        string `json:"client_id"`
	DisplayName     string `json:"display_name"`
	RealName        string `json:"real_name"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	Sex             *string `json:"sex"`
	DefaultEmail    string `json:"default_email"`
	Emails          []string `json:"emails"`
	Birthday        string `json:"birthday"`
	DefaultAvatarID string `json:"default_avatar_id"`
	IsAvatarEmpty   bool `json:"is_avatar_empty"`
	DefaultPhone    struct {
		ID    int    `json:"id"`
		Number string `json:"number"`
	} `json:"default_phone"`
	PSUID string `json:"psuid"`
}

// CallbackHandler returns an HTTP handler for the callback URL
func (y *Yandex) CallbackHandler(c *gin.Context) {
	code := c.Request.URL.Query().Get("code")
	fmt.Println("code: ", code)
	if code == "" {
		http.Error(c.Writer, "Missing authorization code", http.StatusBadRequest)
		return
	}

	tokenURLParams := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"client_id":     {y.config.ClientID},
		"client_secret": {y.config.ClientSecret},
	}
	req, err := http.NewRequest("POST", "https://oauth.yandex.ru/token", bytes.NewBufferString(tokenURLParams.Encode()))
	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Print the response
	fmt.Println(string(body))

	var authResponse AuthResponse
	err = json.Unmarshal(body, &authResponse)
	if err != nil {
		fmt.Println(err)
		return
	}

	accessToken := authResponse.AccessToken
	fmt.Printf("Access token: %s\n", accessToken)

	userURL := "https://login.yandex.ru/info?format=json&oauth_token=" + accessToken
	resp, err = http.Get(userURL)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userResponse User
	err = json.NewDecoder(resp.Body).Decode(&userResponse)
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}

	dbPath := "./db.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	pot_user, _ := database.GetUserByUsername(db, userResponse.RealName)

	if pot_user.UserID == 0 {
		fmt.Println("No such user!")
		res := database.InsertUser(db, userResponse.RealName, userResponse.DefaultEmail, "")
		if !res {
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
			return
		}
		created_user, _ := database.GetUserByUsername(db, userResponse.RealName)
		// Create yandex user
		var yandex_user m.YandexUserInfo
		yandex_user.Birthday = userResponse.Birthday
		yandex_user.DefaultEmail = userResponse.DefaultEmail
		yandex_user.DefaultAvatarID = userResponse.DefaultAvatarID
		yandex_user.DisplayName = userResponse.DisplayName
		yandex_user.FirstName = userResponse.FirstName
		yandex_user.LastName = userResponse.LastName
		yandex_user.RealName = userResponse.RealName
		yandex_user.YandexID = userResponse.ID
		yandex_user.ID = created_user.UserID
		yandex_user.Number = userResponse.DefaultPhone.Number

		_, err = database.CreateYandexUser(db, yandex_user)
		if err != nil {
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
			return
		}

		// Create wallet
		err = database.CreateWallet(db, yandex_user.ID)
		if err != nil {
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
			return
		}

		wallet, err := database.GetBalanceByUserID(db, yandex_user.ID)
		if err != nil {
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return

		c.SetCookie(
			"username",
			userResponse.RealName,
			3600,
			"/",
			"localhost",
			false,
			true,
		)
		c.HTML(
			http.StatusOK,
			"articles/account.html",
			gin.H{
				"current_user": yandex_user.RealName,
				"username":     yandex_user.RealName,
				"email":        yandex_user.DefaultEmail,
				"id":           yandex_user.ID,
				"balance":      wallet,
				"avatar_url":   yandex_user.DefaultAvatarID,
			},
		)

	} else {
		yandex_user, err := database.GetUserByUsername(db, userResponse.RealName)
		if err != nil {
			http.Error(c.Writer, err.Error()+"253", http.StatusInternalServerError)
			return
		}

		avatarURL, err := database.GetYandexAvatarURLByUsername(db, userResponse.RealName)
		if err != nil {
			http.Error(c.Writer, err.Error()+"259", http.StatusInternalServerError)
			return
		}

		wallet, err := database.GetBalanceByUserID(db, yandex_user.UserID)
		if err != nil {
			http.Error(c.Writer, err.Error()+"265", http.StatusInternalServerError)
			return
		}

		c.SetCookie(
			"username",
			userResponse.RealName,
			3600,
			"/",
			"localhost",
			false,
			true,
		)
		c.HTML(
			http.StatusOK,
			"articles/account.html",
			gin.H{
				"current_user": yandex_user.Username,
				"username":     yandex_user.Username,
				"email":        yandex_user.Email,
				"id":           yandex_user.UserID,
				"balance":      wallet,
				"avatar_url":   avatarURL,
			},
		)
	}
}