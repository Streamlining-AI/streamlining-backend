package middleware

import (
	"fmt"
	"net/http"

	helper "github.com/Streamlining-AI/streamlining-backend/helpers"

	"github.com/gin-gonic/gin"
)

// Authz validates token and authorizes users
func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientToken, err := c.Cookie("token")
		if clientToken == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("No Authorization header provided")})
			c.Abort()
			return
		}
		claims, errr := helper.ValidateToken(clientToken)
		if errr != "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			c.Abort()
			return
		}

		c.Set("email", claims.Email)
		c.Set("uid", claims.Uid)

		c.Next()

	}
}

func AuthHandler(c *gin.Context) {
	// Get the token from the cookie
	token, err := c.Cookie("token")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Verify the token and get the user details
	claims, msg := helper.DecodeToken(token)
	if msg != "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": msg})
		return
	}
	c.Set("user", claims)

	// Call the next middleware function
	c.Next()
	// Use the user details to authenticate the user
	// ...
	// Return a success message
	c.JSON(http.StatusOK, gin.H{"message": "Authentication successful!"})
}
