package middleware

import (
	"fmt"
	"gorm/initializers"
	"gorm/models"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func RequireAuth(c *gin.Context) {
	fmt.Println("In middleware")

	// Get the cookie off req
	tokenString, err := c.Cookie("Authorization")

	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
	// Decode/ validate it

	// Parse takes the token string and a function for looking up the key
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		//Dont forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		//hmacSampleSecret is a []byte containing your secret, e.g. []byte
		return []byte(os.Getenv("SECRET")), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

		// Check the exp
		if float64(time.Now().Unix()) > claims["expiration"].(float64) {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
		// Find the user with token sub
		var user models.User
		initializers.DB.First(&user, claims["subject"])

		if user.ID == 0 {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

		// Attach to req
		c.Set("user", user)

		// Continue
		c.Next()

	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
	}

}
