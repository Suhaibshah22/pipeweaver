package cmd

import (
	"context"
	"os"

	"github.com/Suhaibshah22/pipeweaver/cmd/config"
	"github.com/Suhaibshah22/pipeweaver/cmd/controller"
	"github.com/Suhaibshah22/pipeweaver/external"
	"github.com/Suhaibshah22/pipeweaver/internal/adapter/repository"
	port "github.com/Suhaibshah22/pipeweaver/internal/domain/port/repository"
	"github.com/Suhaibshah22/pipeweaver/internal/usecase"

	"log/slog"

	"github.com/Suhaibshah22/pipeweaver/util"
)

type Container struct {
	// Logger
	Logger *slog.Logger

	// Config
	Config *config.Config

	// Repositories
	GitRepository port.GitRepository

	// Usecases
	ProcessRepositoryUseCase  usecase.ProcessPipelineUsecase
	GenerateAirFlowDAGUsecase usecase.GenerateAirFlowDAGUsecase

	// Controllers
	WebhookController *controller.WebhookController

	// External Services
	GitService    external.GitService
	GitHubService external.GitHubService
}

func InitializeContainer(cfg *config.Config, ctx context.Context) *Container {
	container := &Container{
		Config: cfg,
	}

	// Initialize Logger
	util.InitLogger(cfg.LogLevel(), cfg.Environment())
	container.Logger = util.GetLogger()

	// Initialize Git Repository
	gitRepo, err := repository.NewGitRepository(
		cfg.Git.RemoteURL,
		cfg.Git.DefaultBranch,
		cfg.App.RepoBaseDir,
		cfg.Git.Username,
		cfg.Git.Token,
	)
	if err != nil {
		container.Logger.Error("Failed to initialize Git Repository", "error", err)
		os.Exit(1)
	}
	container.GitRepository = gitRepo

	// Initialize External Services
	container.GitService = external.NewGitService()
	container.GitHubService = external.NewGitHubService()

	// Initialize Usecases
	container.GenerateAirFlowDAGUsecase = usecase.NewGenerateAirFlowDAGUsecase(container.Logger)
	container.ProcessRepositoryUseCase = usecase.NewProcessPipelineUsecase(
		container.GitRepository,
		container.GitHubService,
		container.GenerateAirFlowDAGUsecase,
		container.Logger,
		cfg)

	// Initialize Controllers
	container.WebhookController = controller.NewWebhookController(container.ProcessRepositoryUseCase, container.Logger, cfg)

	// Start the queue worker in a separate goroutine
	go func() {
		container.ProcessRepositoryUseCase.StartQueue(ctx)
	}()

	return container
}
