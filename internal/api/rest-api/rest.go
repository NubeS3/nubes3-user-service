package rest_api

import (
	"github.com/Nubes3/nubes3-user-service/internal/aggregate"
	"github.com/Nubes3/nubes3-user-service/internal/api/middlewares"
	"github.com/gin-gonic/gin"
	"net/http"
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
		userRoutesGroup.POST("/signin", aggregate.SignInHandler)

		userRoutesGroup.POST("/signup", aggregate.SignUpHandler)

		userRoutesGroup.PUT("/resend-otp", aggregate.ResendOtpHandler)

		userRoutesGroup.POST("/confirm-otp", aggregate.ConfirmOtpHandler)

		userRoutesGroup.POST("/update", middlewares.UserAuthenticate, aggregate.UpdateUserHandler)
	}
}

//func sendOTP(username string, email string, otp string, expiredTime time.Time) error {
//	return nats.SendEmailEvent(email, username, otp, expiredTime)
//}
