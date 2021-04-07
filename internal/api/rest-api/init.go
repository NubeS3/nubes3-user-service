package rest_api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Routing(r *gin.Engine) {

}

func Servee() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, "pong")
	})
	Routing(r)

	_ = r.Run(":6160")
}
