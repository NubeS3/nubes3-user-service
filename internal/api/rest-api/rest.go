package rest_api

import (
	"github.com/Nubes3/nubes3-user-service/internal/api/middlewares"
	"github.com/Nubes3/nubes3-user-service/internal/repo/arangodb"
	"github.com/Nubes3/nubes3-user-service/pkg/models"
	"github.com/Nubes3/nubes3-user-service/pkg/utils"
	scrypt "github.com/elithrar/simple-scrypt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func routing(r *gin.Engine) {

}

func Serve() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, "pong")
	})
	routing(r)

	_ = r.Run(":6160")
}

func UserRoutes(route *gin.Engine) {
	userRoutesGroup := route.Group("/users")
	{
		userRoutesGroup.POST("/signin", func(c *gin.Context) {
			type signinUser struct {
				Username string `json:"username" binding:"required"`
				Password string `json:"password" binding:"required"`
			}
			var curSigninUser signinUser
			if err := c.ShouldBind(&curSigninUser); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
				return
			}

			user, err := arangodb.FindUserByUsername(curSigninUser.Username)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "invalid username",
				})
				return
			}

			err = scrypt.CompareHashAndPassword([]byte(user.Pass), []byte(curSigninUser.Password))
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": err.Error(),
				})
				return
			}

			if !user.IsActive {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "user have not verified account via otp",
				})
				return
			}

			accessToken, err := utils.CreateToken(user.Id)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
				//_ = nats.SendErrorEvent(err.Error()+" at user route/sign in/access token",
				//	"Token Error")
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"accessToken":  accessToken,
				"refreshToken": user.RefreshToken,
			})
		})

		userRoutesGroup.POST("/signup", func(c *gin.Context) {
			var user models.User
			if err := c.ShouldBind(&user); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
				return
			}

			_, err := arangodb.CreateUser(
				user.Firstname,
				user.Lastname,
				user.Username,
				user.Pass,
				user.Email,
				user.Dob,
				user.Company,
				user.Gender,
			)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
				return
			}

			utils.SendOtp(user.Firstname+" "+user.Lastname, user.Username, user.Email, user.Otp.Val, user.Otp.Expired)

			c.JSON(http.StatusOK, gin.H{
				"message": "verify account via otp sent to your email",
			})
		})

		userRoutesGroup.PUT("/resend-otp", func(c *gin.Context) {
			type resendUser struct {
				Username string `json:"username"`
			}
			var requestedUser resendUser
			if err := c.ShouldBind(&requestedUser); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
			}

			user, err := arangodb.UpdateOtp(requestedUser.Username)
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "username not found",
				})
			}

			utils.SendOtp(user.Firstname+" "+user.Lastname, user.Username, user.Email, user.Otp.Val, user.Otp.Expired)

			c.JSON(http.StatusOK, gin.H{
				"message": "otp resent",
			})

		})

		userRoutesGroup.POST("/confirm-otp", func(c *gin.Context) {
			type otpValidate struct {
				Username string `json:"username"`
				Otp      string `json:"otp"`
			}
			var curSigninUser otpValidate
			if err := c.ShouldBind(&curSigninUser); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
				return
			}

			_, err := arangodb.ConfirmOtp(curSigninUser.Username, curSigninUser.Otp)
			if err != nil {
				if merr, ok := err.(*utils.ModelError); ok {
					if merr.ErrType == utils.NotFound {
						c.JSON(http.StatusNotFound, gin.H{
							"error": "otp expired or mismatch",
						})

						return
					}

					if merr.ErrType == utils.DbError {
						c.JSON(http.StatusInternalServerError, gin.H{
							"error": "something went wrong",
						})

						//_ = nats.SendErrorEvent(err.Error()+" at user route/confirm otp/models otp confirm",
						//	"Db Error")
						return
					}
				}

				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "something went wrong",
				})
				//_ = nats.SendErrorEvent(err.Error()+" at user route/confirm otp/models otp confirm",
				//	"Db Error")
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"message": "otp confirmed",
			})
		})

		userRoutesGroup.POST("/update", middlewares.UserAuthenticate, func(c *gin.Context) {
			type updateUser struct {
				Firstname string    `json:"firstname" binding:"required"`
				Lastname  string    `json:"lastname" binding:"required"`
				Dob       time.Time `json:"dob" binding:"required"`
				Company   string    `json:"company" binding:"required"`
				Gender    bool      `json:"gender" binding:"required"`
			}

			var curUpdateUser updateUser
			if err := c.ShouldBind(&curUpdateUser); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": err.Error(),
				})
				return
			}

			uid, ok := c.Get("uid")
			if !ok {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "something when wrong",
				})

				//_ = nats.SendErrorEvent("uid not found in authenticate at /users/update",
				//	"Unknown Error")
				return
			}

			user, err := arangodb.UpdateUserData(uid.(string), curUpdateUser.Firstname, curUpdateUser.Lastname,
				curUpdateUser.Dob, curUpdateUser.Company, curUpdateUser.Gender)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "something when wrong",
				})

				//_ = nats.SendErrorEvent(err.Error()+" at authenticated users/update",
				//	"Db Error")
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"firstname": user.Firstname,
				"lastname":  user.Lastname,
				"dob":       user.Dob,
				"company":   user.Company,
				"gender":    user.Gender,
			})
		})
	}
}

//func sendOTP(username string, email string, otp string, expiredTime time.Time) error {
//	return nats.SendEmailEvent(email, username, otp, expiredTime)
//}
