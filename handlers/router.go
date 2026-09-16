package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/infracloudio/msbotbuilder-go/controllers"
)

// NewRouter initializes and returns the Gin engine router for the proactive bot service
func NewRouter(botCtrl *controllers.BotController, alertCtrl *controllers.AlertController) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	router.POST("/api/messages", gin.WrapF(botCtrl.HandleMessage))
	router.POST("/api/alerts", gin.WrapF(alertCtrl.SendAlert))
	router.GET("/api/sessions", gin.WrapF(alertCtrl.GetSessions))

	return router
}
