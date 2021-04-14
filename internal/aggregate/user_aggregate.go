package aggregate

import (
	"encoding/json"
	models "github.com/Nubes3/common/models/arangodb"
	"github.com/Nubes3/common/models/nats"
	"github.com/Nubes3/common/utils"
	"github.com/Nubes3/nubes3-user-service/internal/repo/arangodb"
	repo "github.com/Nubes3/nubes3-user-service/internal/repo/arangodb"
	"github.com/dgrijalva/jwt-go"
	scrypt "github.com/elithrar/simple-scrypt"
	"github.com/gin-gonic/gin"
	n "github.com/nats-io/nats.go"
	"github.com/prometheus/common/log"
	"net/http"
	"time"
)

func SignInHandler(c *gin.Context) {
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
}

func SignUpHandler(c *gin.Context) {
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
}

func ResendOtpHandler(c *gin.Context) {
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

}

func ConfirmOtpHandler(c *gin.Context) {
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
}

func UpdateUserHandler(c *gin.Context) {
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
}

func UserMessageRequestHandler(msg *n.Msg) {
	message := nats.Msg{}
	err := json.Unmarshal(msg.Data, &message)
	if err != nil {
		//TODO log
		log.Error("Unknown format: " + string(msg.Data))
	}

	if message.ReqType == nats.GetById {
		id := message.Data

		response := nats.MsgResponse{}

		user, err := repo.FindUserById(id)
		if err != nil {
			response.IsErr = true
			jsonData, err := json.Marshal(err.(*utils.ModelError))
			if err != nil {
				// TODO log error here
			}

			response.Data = string(jsonData)
			resJson, _ := json.Marshal(response)
			_ = msg.Respond(resJson)
			return
		}

		response.IsErr = false
		jsonData, err := json.Marshal(user)
		response.Data = string(jsonData)

		resJson, _ := json.Marshal(response)
		_ = msg.Respond(resJson)
		return
	}

	if message.ReqType == nats.Resolve {
		authToken := message.Data

		response := nats.MsgResponse{}

		var userClaims utils.UserClaims
		token, err := utils.ParseToken(authToken, &userClaims)

		user, userErr := repo.FindUserById(userClaims.Id)
		if userErr != nil {
			jsonData, _ := json.Marshal(utils.ModelError{
				Msg:     "user not found",
				ErrType: utils.NotFound,
			})

			response.Data = string(jsonData)
			resJson, _ := json.Marshal(response)
			_ = msg.Respond(resJson)
			return
		}

		if err != nil {
			response.IsErr = true
			validationError, _ := err.(*jwt.ValidationError)

			if validationError.Errors == jwt.ValidationErrorExpired {
				rfToken := message.ExtraData[0]
				if rfToken == "" {
					jsonData, _ := json.Marshal(utils.ModelError{
						Msg:     "token expired",
						ErrType: utils.Expired,
					})

					response.Data = string(jsonData)
					resJson, _ := json.Marshal(response)
					_ = msg.Respond(resJson)
					return
				}

				if user.RefreshToken != rfToken {
					jsonData, _ := json.Marshal(utils.ModelError{
						Msg:     "token expired",
						ErrType: utils.Expired,
					})

					response.Data = string(jsonData)
					resJson, _ := json.Marshal(response)
					_ = msg.Respond(resJson)
					return
				}

				user, err = repo.UpdateRefreshToken(user.Id)
				if err != nil {
					jsonData, _ := json.Marshal(utils.ModelError{
						Msg:     "token expired",
						ErrType: utils.Expired,
					})

					response.Data = string(jsonData)
					resJson, _ := json.Marshal(response)
					_ = msg.Respond(resJson)
					//TODO log error
					return
				}

				newAccessToken, err := utils.CreateToken(user.Id)
				if err != nil {
					jsonData, _ := json.Marshal(utils.ModelError{
						Msg:     "token expired",
						ErrType: utils.Expired,
					})

					response.Data = string(jsonData)

					//TODO log error
					//_ = nats.SendErrorEvent(err.Error()+" at user route/sign in/access token",
					//	"Token Error")
					return
				}

				response.Data = userClaims.Id
				response.ExtraData = []string{newAccessToken, user.RefreshToken}

				return
			}

			if err == jwt.ErrSignatureInvalid {
				jsonData, _ := json.Marshal(utils.ModelError{
					Msg:     "token invalid",
					ErrType: utils.Invalid,
				})

				response.Data = string(jsonData)
				return
			} else {
				jsonData, _ := json.Marshal(utils.ModelError{
					Msg:     "token invalid",
					ErrType: utils.Invalid,
				})

				response.Data = string(jsonData)
				resJson, _ := json.Marshal(response)
				_ = msg.Respond(resJson)
				return
			}
		}

		if !token.Valid {
			jsonData, _ := json.Marshal(utils.ModelError{
				Msg:     "token invalid",
				ErrType: utils.Invalid,
			})

			response.Data = string(jsonData)
			resJson, _ := json.Marshal(response)
			_ = msg.Respond(resJson)
			return
		}

		userJson, _ := json.Marshal(user)
		response.Data = string(userJson)
		resJson, _ := json.Marshal(response)
		_ = msg.Respond(resJson)
	}
}
