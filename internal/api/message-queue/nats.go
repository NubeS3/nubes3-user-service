package message_queue

import (
	"encoding/json"
	"github.com/Nubes3/common/models/nats"
	"github.com/Nubes3/common/utils"
	repo "github.com/Nubes3/nubes3-user-service/internal/repo/arangodb"
	"github.com/dgrijalva/jwt-go"
	n "github.com/nats-io/nats.go"
	"github.com/prometheus/common/log"
)

var sub *n.Subscription

func CreateMessageSubcribe() (func(), error) {
	var err error
	sub, err = nats.Nc.QueueSubscribe(nats.UserSubj, "user_nubes3_q", func(msg *n.Msg) {
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
	})

	if err != nil {
		return nil, err
	}
	return cleanup, nil
}

func cleanup() {
	_ = sub.Unsubscribe()
}
