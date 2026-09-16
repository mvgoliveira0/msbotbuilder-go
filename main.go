package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/infracloudio/msbotbuilder-go/config"
	"github.com/infracloudio/msbotbuilder-go/controllers"
	"github.com/infracloudio/msbotbuilder-go/core"
	"github.com/infracloudio/msbotbuilder-go/handlers"
	"github.com/infracloudio/msbotbuilder-go/repository"
	"github.com/infracloudio/msbotbuilder-go/service"
)

func main() {
	// 1. Load Config
	cfg := config.LoadConfig()

	// 2. Adapter Settings
	adapterSetting := core.AdapterSetting{
		AppID:             cfg.AppID,
		AppPassword:        cfg.AppPassword,
		ChannelAuthTenant: cfg.ChannelAuthTenant,
	}

	adapter, err := core.NewBotAdapter(adapterSetting)
	if err != nil {
		log.Fatal("Error creating bot adapter: ", err)
	}

	// 3. Initialize Repository
	repo := repository.NewInMemoryConversationRepository()

	// 4. Initialize Service
	alertSvc := service.NewAlertService(adapter, repo)

	// 5. Initialize Controller
	alertCtrl := controllers.NewAlertController(alertSvc)

	// 6. Initialize Router
	router := handlers.NewRouter(alertCtrl)

	// 7. Start HTTP Server
	addr := ":" + cfg.Port
	fmt.Printf("Starting proactive bot server on port %s...\n", cfg.Port)
	log.Fatal(http.ListenAndServe(addr, router))
}
