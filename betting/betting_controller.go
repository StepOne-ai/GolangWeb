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
	Name string `form:"name" binding:"required"`
	Group string `form:"group" binding:"required"`
}

func BettingPost(c *gin.Context) {
	var data FormData

	c.Bind(&data);

	if data.Name == "" || data.Group == "" {
		c.Redirect(302, "/betting")
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
	}

	candidate, err := database.GetCandidateByName(db, data.Name)

	// Handle error
	if err != nil {
		c.Redirect(302, "/betting")
	}
	
	c.HTML(http.StatusOK, "articles/candidate.html", gin.H{
		"CandidateID": candidate.CandidateID,
		"Name": candidate.Name,
		"Group": candidate.Group,
		"UpVotes": candidate.UpVotes,
		"DownVotes": candidate.DownVotes,
		"username": c.MustGet("username").(string),
	})
}

func VoteWin(c *gin.Context) {
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

	if vote.VoteID == 0 && vote.UserID == 0 && vote.CandidateID == 0 && vote.VoteType == "" {
		database.RegisterVote(db, user_id, id, "win")
	}

	c.Redirect(302, "/betting")
}

func VoteLose(c *gin.Context) {
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

	if vote.VoteID == 0 && vote.UserID == 0 && vote.CandidateID == 0 && vote.VoteType == "" {
		database.RegisterVote(db, user_id, id, "lose")
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
		database.ClearVote(db, vote.VoteID)
	}

	c.Redirect(302, "/betting")
}