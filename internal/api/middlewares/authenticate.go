package middlewares

import (
	"github.com/Nubes3/common/utils"
	"github.com/Nubes3/nubes3-user-service/internal/repo/arangodb"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func UserAuthenticate(c *gin.Context) {
	authToken := c.GetHeader("Authorization")
	auths := strings.Split(authToken, "Bearer ")
	if len(auths) < 2 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "unauthorized",
		})
		c.Abort()
		return
	}

	authToken = auths[1]
	var userClaims utils.UserClaims
	token, err := utils.ParseToken(authToken, &userClaims)

	if err != nil {
		validationError, _ := err.(*jwt.ValidationError)

		if validationError.Errors == jwt.ValidationErrorExpired {
			rfToken := c.GetHeader("Refresh")
			if rfToken == "" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "please log in again",
				})
				c.Abort()
				return
			}

			user, err := arangodb.FindUserById(userClaims.Id)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "unauthorized",
				})
				c.Abort()
				return
			}

			if user.RefreshToken != rfToken {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "unauthorized",
				})
				c.Abort()
				return
			}

			user, err = arangodb.UpdateRefreshToken(user.Id)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "unauthorized",
				})
				c.Abort()
				return
			}

			newAccessToken, err := utils.CreateToken(user.Id)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
				//_ = nats.SendErrorEvent(err.Error()+" at user route/sign in/access token",
				//	"Token Error")
				c.Abort()
				return
			}

			c.Writer.Header().Set("AccessToken", newAccessToken)
			c.Writer.Header().Set("RefreshToken", user.RefreshToken)
			c.Set("uid", userClaims.Id)
			c.Next()
			return
		}

		if err == jwt.ErrSignatureInvalid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			c.Abort()
			return
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "access key invalid",
			})
			c.Abort()
			return
		}
	}

	if !token.Valid {
		if err == jwt.ErrSignatureInvalid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			c.Abort()
			return
		}
	}

	c.Set("uid", userClaims.Id)
	c.Next()
}
