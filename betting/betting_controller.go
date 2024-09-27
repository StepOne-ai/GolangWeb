package betting

import (
	"database/sql"
	"dbgolang/database"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

func BettingIndex(c *gin.Context) {
	username, err := c.Cookie("username")

	if err != nil {
		c.Redirect(302, "/")
	}

	adminAccess, err := c.Cookie("adminAccess")
	if err != nil {
		adminAccess = "false"
	}

	// Get all candidates
	dbPath := "./db.db"
    db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	candidates, err := database.GetAllCandidates(db)
	if err != nil {
		log.Fatal(err)
	}

	c.HTML(http.StatusOK, "articles/betting.html", gin.H{
		"username": username,
		"candidates": candidates,
		"adminAccess": adminAccess,
	})
}

type FormData struct {
	Name string `form:"name"`
	Group string `form:"group"`
}

func BettingPost(c *gin.Context) {
	var data FormData

	c.Bind(&data);

	if data.Name == "" || data.Group == "" {
		c.Redirect(302, "/betting")
		return
	}

	dbPath := "./db.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	_, err = database.CreateNewCandidate(db, data.Name, data.Group)
	// Handle error
	if err != nil {
		c.Redirect(302, "/betting")
		return
	}

	candidate, err := database.GetCandidateByName(db, data.Name)

	// Handle error
	if err != nil {
		c.Redirect(302, "/betting")
		return
	}
	
	c.HTML(http.StatusOK, "articles/candidate.html", gin.H{
		"CandidateID": candidate.CandidateID,
		"Name": candidate.Name,
		"Group": candidate.Group,
		"UpVotes": candidate.UpVotes,
		"DownVotes": candidate.DownVotes,
	})
}

type Amount struct {
	Amount int `form:"amount"`
}

func VoteWin(c *gin.Context) {
	var data Amount
	if err := c.Bind(&data); err != nil {
		c.Redirect(302, "/")
		return
	}
	value := data.Amount
	fmt.Println("value: ", value)
	
	dbPath := "./db.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()
	// Check if enough
	username, err := c.Cookie("username")
	if err != nil {
		c.Redirect(302, "/")
		return
	}
	user_id, err := database.GetUserIdByUsername(db, username)
	if err != nil {
		c.Redirect(302, "/")
		return
	}
	balance, err := database.GetBalanceByUserID(db, user_id)
	// Handle error
	fmt.Println("balance: ", balance)
	if err != nil {
		c.Redirect(302, "/")
		return
	}

	if balance > value {
		fmt.Println("balance>value")
		// Get the candidate id from the url
		candidateID := c.Param("id")
		id, _ := strconv.Atoi(candidateID)

		vote, _ := database.GetVoteByUserAndCandidate(db, user_id, id)

		fmt.Println("vote.VoteID: ", vote.VoteID)
		if vote.VoteID == 0 && value > 0 {
			fmt.Println("Register vote")
			database.RegisterVote(db, user_id, id, "win", value)
		} else {
			c.Redirect(302, "/betting")
			return
		}

		database.UpdateBalance(db, user_id, -value)

		c.Redirect(302, "/betting")
		return
	} else {
		c.HTML(http.StatusOK, "articles/error.html", gin.H{
			"error": "Insufficient balance!"})

		return
	}
}

func VoteLose(c *gin.Context) {
	var data Amount
	c.Bind(&data)
	value := data.Amount
	// Get the candidate id from the url
	candidateID := c.Param("id")
	id, _ := strconv.Atoi(candidateID)
	fmt.Println("id: ", id)

	// Get user
	username, err := c.Cookie("username")
	if err != nil {
		c.Redirect(302, "/")
	}
	dbPath := "./db.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()
	// Get vote if exists
	user_id, err := database.GetUserIdByUsername(db, username)

	if err != nil {
		c.Redirect(302, "/betting")
	}

	vote, _ := database.GetVoteByUserAndCandidate(db, user_id, id)

	if vote.VoteID == 0 && vote.UserID == 0 && vote.CandidateID == 0 && vote.VoteType == "" {
		balance, err := database.GetBalanceByUserID(db, user_id)
		if err != nil {
			c.HTML(http.StatusOK, "articles/error.html", gin.H{
				"error": "Error getting balance!"})
			return
		}
		if balance > value {
			database.UpdateBalance(db, user_id, -value)
			database.RegisterVote(db, user_id, id, "lose", value)
		} else {
			c.HTML(http.StatusOK, "articles/error.html", gin.H{
				"error": "Insufficient balance!"})
			return
		}
	}

	c.Redirect(302, "/betting")
}

func VoteClear(c *gin.Context) {
	// Get the candidate id from the url
	candidateID := c.Param("id")
	fmt.Println("id: ", candidateID)
	id, _ := strconv.Atoi(candidateID)
	fmt.Println("id: ", id)

	// Get user
	username, err := c.Cookie("username")
	if err != nil {
		c.Redirect(302, "/")
	}

	dbPath := "./db.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()
	// Get vote if exists
	user_id, err := database.GetUserIdByUsername(db, username)

	if err != nil {
		c.Redirect(302, "/betting")
	}

	vote, err := database.GetVoteByUserAndCandidate(db, user_id, id)
	if err != nil {
		c.Redirect(302, "/betting")
	}

	if vote.VoteID != 0 {
		database.UpdateBalance(db, user_id, vote.Amount)
		database.ClearVote(db, vote.VoteID)
	}

	c.Redirect(302, "/betting")
}