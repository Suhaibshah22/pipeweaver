package controller

import (
	"context"
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

	// Only process the repository if the event is a push to the main branch
	if payload.Ref != "refs/heads/main" {
		wc.Log.Info("Ignoring event", "event", payload.Ref)
		c.JSON(http.StatusOK, gin.H{"status": "Ignoring event"})
		return
	}

	// Process the repository using the use case
	err = wc.ProcessPipelineUsecase.Execute(context.Background(), payload)
	if err != nil {
		wc.Log.Error("Failed to process repository", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process repository"})
		return
	}

	// Respond with success
	c.JSON(http.StatusOK, gin.H{"status": "Repository processed successfully"})
}
