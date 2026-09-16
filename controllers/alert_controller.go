package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/infracloudio/msbotbuilder-go/models"
	"github.com/infracloudio/msbotbuilder-go/service"
)

// AlertController handles HTTP requests for Bot Framework webhook, proactive alerts, and sessions
type AlertController struct {
	alertService service.AlertService
}

// NewAlertController creates a new AlertController instance
func NewAlertController(alertService service.AlertService) *AlertController {
	return &AlertController{alertService: alertService}
}

// HandleMessage handles POST /api/messages (incoming Bot Framework webhook)
func (c *AlertController) HandleMessage(w http.ResponseWriter, req *http.Request) {
	err := c.alertService.ProcessWebhookRequest(req)
	if err != nil {
		fmt.Println("[AlertController] Error processing webhook request:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
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

	err = c.alertService.SendProactiveAlert(req.Context(), reqBody.TenantID, reqBody.UserID, reqBody.Message)
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
	sessions, err := c.alertService.GetActiveSessions(req.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(sessions)
}
