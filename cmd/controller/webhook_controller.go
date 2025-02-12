package controller

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/Suhaibshah22/pipeweaver/internal/usecase"

	"github.com/Suhaibshah22/pipeweaver/external"

	"github.com/Suhaibshah22/pipeweaver/cmd/config"

	"github.com/gin-gonic/gin"
)

type WebhookController struct {
	ProcessPipelineUsecase usecase.ProcessPipelineUsecase
	Log                    *slog.Logger
	Config                 *config.Config
}

func NewWebhookController(
	ProcessPipelineUsecase usecase.ProcessPipelineUsecase,
	logger *slog.Logger,
	cfg *config.Config,
) *WebhookController {
	return &WebhookController{
		ProcessPipelineUsecase: ProcessPipelineUsecase,
		Log:                    logger,
		Config:                 cfg,
	}
}

func (wc *WebhookController) HandleWebhook(c *gin.Context) {
	var payload external.GitHubWebhookPayload

	// Read the body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		wc.Log.Error("Error reading request body", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Parse JSON
	if err := json.Unmarshal(body, &payload); err != nil {
		wc.Log.Error("Error parsing JSON", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Enqueue Processing the request using the usecase
	select {
	case usecase.ProcessPipelinesQueue <- payload:
		wc.Log.Info("Webhook request enqueued", "repo", payload.Repository.Name)
		c.JSON(http.StatusAccepted, gin.H{"status": "Webhook request accepted for processing"})
	default:
		// Queue is full, reject request
		wc.Log.Error("Queue is full, dropping request")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Queue is full, try again later"})
	}
}
