package models

type User struct {
	UserID int
	Username string
	PasswordHash string
	Email string
}

type Article struct {
	ArticleID int
	Title string
	Content string
	Author string
}

type Candidate struct {
	CandidateID int
	Name string
	Group string
	UpVotes int
	DownVotes int
}

type Group struct {
	GroupID int
	Name string
	Candidates []Candidate
}

type Vote struct {
	VoteID int
	UserID int
	CandidateID int
	VoteType string
}