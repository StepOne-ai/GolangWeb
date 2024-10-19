package vk

import (
	"database/sql"
	"dbgolang/database"
	"dbgolang/models"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/matthewhartstonge/pkce"
)

const (
	vkClientID     = "52467139"
	vkClientSecret = "a75272fe575272fe575272fe55d7607ba267752775272fe5122d60bf2260bbcefc4e0f43"
	vkRedirectURI  = "http://localhost/callback"
)

var verifier string
var challenge string

func setChallengeAndVerifier() {
	key, err := pkce.New()
	if err != nil {
		log.Fatal(err)
	}
	verifier = key.CodeVerifier()
	challenge = key.CodeChallenge()
}


func VkDetails(c *gin.Context) {
	// Get username
	username := c.Params.ByName("username")
	fmt.Println(username)

	c.HTML(
		http.StatusOK,
		"articles/vk.html",
		gin.H{},
	)
}

func VkRegister(c *gin.Context) {
	setChallengeAndVerifier()
	url_for_code := "https://id.vk.com/authorize?response_type=code&client_id=52467139&scope=email%20phone&redirect_uri=http://localhost/callback&state=state&code_challenge="+challenge+"&code_challenge_method=s256"

	http.Redirect(c.Writer, c.Request, url_for_code, http.StatusFound)
}

func VkLogin(c *gin.Context) {
	setChallengeAndVerifier()
	url_for_code := "https://id.vk.com/authorize?response_type=code&client_id=52467139&scope=email%20phone&redirect_uri=http://localhost/callback&state=state&code_challenge="+challenge+"&code_challenge_method=s256"

	http.Redirect(c.Writer, c.Request, url_for_code, http.StatusFound)
}

func CallbackHandler(c *gin.Context) {
	code := c.Query("code")
	device_id := c.Query("device_id")
	
	fmt.Println("Code: ", code)
	fmt.Println("Device_id: ", device_id)

	response, err := getAccessToken(code, device_id)
	if err != nil {
		log.Fatal(err)
	}

	user_id := response.User_id
	access_token := response.Access_token
	fmt.Println("User_id: ", user_id)
	fmt.Println("Access_token: ", access_token)

	type User struct {
		ID           int    `json:"id"`
		BDate        string `json:"bdate"`
		Photo200Orig string `json:"photo_200_orig"`
		Interests    string `json:"interests"`
		About        string `json:"about"`
		Activities   string `json:"activities"`
		University   int    `json:"university"`
		UniversityName string `json:"university_name"`
		Faculty      int    `json:"faculty"`
		FacultyName  string `json:"faculty_name"`
		Graduation   int    `json:"graduation"`
		HomeTown     string `json:"home_town"`
		Personal     struct {
			InspiredBy string `json:"inspired_by"`
		} `json:"personal"`
		Schools      []interface{} `json:"schools"`
		Sex          int           `json:"sex"`
		FirstName    string        `json:"first_name"`
		LastName     string        `json:"last_name"`
		CanAccessClosed bool        `json:"can_access_closed"`
		IsClosed     bool          `json:"is_closed"`
	}
	type Response struct {
		Response []User `json:"response"`
	}

	// Construct the URL
	url_info := fmt.Sprintf("https://api.vk.com/method/users.get?user_id=%s&fields=about,bdate,city,education,sex,interests,crop_photo,activities,home_town,photo_200_orig,schools,ocupation,personal&name_case=nom&access_token=%s&v=5.199", strconv.Itoa(int(user_id)), access_token)

	// Create a new HTTP request
	req, err := http.NewRequest("POST", url_info, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

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
	var userResponse Response
	// Unmarshal the JSON data into a struct
	err = json.Unmarshal(body, &userResponse)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Print the response
	fmt.Println(userResponse.Response[0].FirstName)

	dbPath := "./db.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	pot_user, _ := database.GetUserByUsername(db, string(userResponse.Response[0].FirstName + " " + userResponse.Response[0].LastName))
	if pot_user.UserID == 0 {
		if database.InsertUser(db, 
			string(userResponse.Response[0].FirstName + " " + userResponse.Response[0].LastName), 
			"Пусто", 
			"") {
				user_id, _ := database.GetUserIdByUsername(db, string(userResponse.Response[0].FirstName + " " + userResponse.Response[0].LastName))
				schoolsStr := ""
				for _, school := range userResponse.Response[0].Schools {
					schoolsStr += fmt.Sprintf("%v, ", school)
				}
				database.CreateVkUser(db, models.VkUserInfo{
					ID:           user_id,
					BDate:        userResponse.Response[0].BDate,
					Photo200Orig: userResponse.Response[0].Photo200Orig,
					Interests:    userResponse.Response[0].Interests,
					About:        userResponse.Response[0].About,
					Activities:   userResponse.Response[0].Activities,
					University:   userResponse.Response[0].University,
					UniversityName: userResponse.Response[0].UniversityName,
					Faculty:      userResponse.Response[0].Faculty,
					FacultyName:  userResponse.Response[0].FacultyName,
					Graduation:   userResponse.Response[0].Graduation,
					HomeTown:     userResponse.Response[0].HomeTown,
					InspiredBy:   userResponse.Response[0].Personal.InspiredBy,
					Schools:      schoolsStr,
					Sex:          userResponse.Response[0].Sex,
					FirstName:    userResponse.Response[0].FirstName,
					LastName:     userResponse.Response[0].LastName,
					CanAccessClosed: userResponse.Response[0].CanAccessClosed,
					IsClosed:     userResponse.Response[0].IsClosed,
				})
				vk_user, err := database.GetUserByUsername(db, string(userResponse.Response[0].FirstName + " " + userResponse.Response[0].LastName))
				if err != nil {
					log.Fatal(err)
				}
				database.CreateWallet(db, vk_user.UserID)
				wallet, err := database.GetBalanceByUserID(db, vk_user.UserID)
				if err != nil {
					log.Fatal(err)
				}
				c.SetCookie("username", string(userResponse.Response[0].FirstName + " " + userResponse.Response[0].LastName), 3600, "/", "", false, true)
				c.HTML(
				http.StatusOK,
				"articles/account.html",
					gin.H{
						"current_user": string(userResponse.Response[0].FirstName + " " + userResponse.Response[0].LastName),
						"username": vk_user.Username,
						"email": vk_user.Email,
						"id": vk_user.UserID,
						"balance": wallet,
						"avatar_url": userResponse.Response[0].Photo200Orig,
					},
				)
			}
	} else {
		schoolsStr := ""
		for _, school := range userResponse.Response[0].Schools {
			schoolsStr += fmt.Sprintf("%v, ", school)
		}
		vk_user, err := database.GetUserByUsername(db, string(userResponse.Response[0].FirstName + " " + userResponse.Response[0].LastName))
		if err != nil {
			log.Fatal(err)
		}
		wallet, err := database.GetBalanceByUserID(db, vk_user.UserID)
		if err != nil {
			log.Fatal(err)
		}
		c.SetCookie("username", string(userResponse.Response[0].FirstName + " " + userResponse.Response[0].LastName), 3600, "/", "", false, true)
		c.HTML(
		http.StatusOK,
		"articles/account.html",
			gin.H{
				"current_user": string(userResponse.Response[0].FirstName + " " + userResponse.Response[0].LastName),
				"username": vk_user.Username,
				"email": vk_user.Email,
				"id": vk_user.UserID,
				"balance": wallet,
				"avatar_url": userResponse.Response[0].Photo200Orig,
			},
		)
	}
}

type TokenResponse struct {
	Refresh_token string `json:"refresh_token"`
	Access_token  string `json:"access_token"`
	Id_token  string `json:"id_token"`
	Token_type  string `json:"token_type"`
	Expires_in  int `json:"expires_in"`
	User_id  int64 `json:"user_id"`
	State string `json:"state"`
	Scope string `json:"scope"`
}

// obtained token.
func getAccessToken(code string, device_id string) (*TokenResponse, error) {

	tokenURL := "https://id.vk.com/oauth2/auth"
	params := url.Values{
		"grant_type":    {"authorization_code"},
		"code_verifier": {verifier},
		"redirect_uri":  {vkRedirectURI},
		"code":          {code},
		"client_id":     {vkClientID},
		"device_id":     {device_id},
		"state":         {"state"},
	}

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(params.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tokenResponse TokenResponse

	err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
	if err != nil {
		return nil, err
	}

	return &tokenResponse, nil
}