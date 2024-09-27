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

type Vote struct {
	VoteID int
	UserID int
	CandidateID int
	VoteType string
	Amount int
}

type Wallet struct {
	WalletID int
	UserID int
	Balance float64
}