package controllers

import (
	"fmt"
	"net/http"

	"github.com/infracloudio/msbotbuilder-go/service"
)

// BotController handles incoming webhook HTTP requests from Bot Framework Connector
type BotController struct {
	botService service.BotService
}

// NewBotController creates a new BotController instance
func NewBotController(botService service.BotService) *BotController {
	return &BotController{botService: botService}
}

// HandleMessage handles POST /api/messages
func (c *BotController) HandleMessage(w http.ResponseWriter, req *http.Request) {
	err := c.botService.ProcessWebhookRequest(req)
	if err != nil {
		fmt.Println("[BotController] Error processing request:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}
