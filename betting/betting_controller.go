package betting 

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"dbgolang/database"
)

func BettingIndex(c *gin.Context) {
	username, err := c.Cookie("username")

	if err != nil {
		c.Redirect(302, "/")
	}

	// Get all groups
	groups, err := database.GetAllGroups()	

	c.HTML(http.StatusOK, "betting/index.html", gin.H{
		"username": username,
	})
}