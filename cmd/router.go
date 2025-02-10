package cmd

import (
	"github.com/gin-gonic/gin"
)

func SetupRouter(container *Container) *gin.Engine {
	router := gin.Default()

	// Webhook Routes
	webhookGroup := router.Group("/webhook")
	{
		webhookGroup.POST("/git", container.WebhookController.HandleWebhook)
	}

	return router
}
