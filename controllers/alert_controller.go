package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/infracloudio/msbotbuilder-go/models"
	"github.com/infracloudio/msbotbuilder-go/service"
)

// AlertController handles REST API requests for triggering proactive alerts
type AlertController struct {
	botService service.BotService
}

// NewAlertController creates a new AlertController instance
func NewAlertController(botService service.BotService) *AlertController {
	return &AlertController{botService: botService}
}

// SendAlert handles POST /api/alerts
func (c *AlertController) SendAlert(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody models.AlertRequest
	err := json.NewDecoder(req.Body).Decode(&reqBody)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = c.botService.SendProactiveAlert(req.Context(), reqBody.TenantID, reqBody.UserID, reqBody.Message)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "Alert sent successfully"})
}

// GetSessions handles GET /api/sessions
func (c *AlertController) GetSessions(w http.ResponseWriter, req *http.Request) {
	sessions, err := c.botService.GetActiveSessions(req.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(sessions)
}
