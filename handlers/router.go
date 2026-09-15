package handlers

import (
	"net/http"

	"github.com/infracloudio/msbotbuilder-go/controllers"
)

// NewRouter initializes and returns the HTTP ServeMux router for the proactive bot service
func NewRouter(botCtrl *controllers.BotController, alertCtrl *controllers.AlertController) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/messages", botCtrl.HandleMessage)
	mux.HandleFunc("/api/alerts", alertCtrl.SendAlert)

	return mux
}
