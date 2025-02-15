package usecase

import (
	"context"
	"crypto/rand"
	"log"
	"log/slog"
	"math/big"
	"path/filepath"
	"strings"

	"github.com/Suhaibshah22/pipeweaver/cmd/config"
	"github.com/Suhaibshah22/pipeweaver/external"
	"github.com/Suhaibshah22/pipeweaver/internal/domain/entity"
	"github.com/Suhaibshah22/pipeweaver/internal/domain/port/repository"
	"github.com/Suhaibshah22/pipeweaver/util"
)

// Corrected paths to reflect the correct structure in the repository
const PIPELINES_DIRECTORY = "pipelines/"
const OUTPUT_DIRECTORY = "airflow-dags/"

var ProcessPipelinesQueue chan external.GitHubWebhookPayload

type ProcessPipelineUsecase interface {
	execute(ctx context.Context, payload external.GitHubWebhookPayload) error
	StartQueue(ctx context.Context)
}

type processPipelineUsecase struct {
	GitRepository             repository.GitRepository
	GitHubService             external.GitHubService
	GenerateAirFlowDAGUsecase GenerateAirFlowDAGUsecase

	Config *config.Config
	Log    *slog.Logger
}

func NewProcessPipelineUsecase(
	gitRepo repository.GitRepository,
	gitHubService external.GitHubService,
	generateAirFlowDAGUsecase GenerateAirFlowDAGUsecase,

	logger *slog.Logger,
	cfg *config.Config,
) ProcessPipelineUsecase {
	ProcessPipelinesQueue = make(chan external.GitHubWebhookPayload, 100)

	return &processPipelineUsecase{
		GitRepository:             gitRepo,
		GitHubService:             gitHubService,
		GenerateAirFlowDAGUsecase: generateAirFlowDAGUsecase,

		Config: cfg,
		Log:    logger,
	}
}

func (uc *processPipelineUsecase) StartQueue(ctx context.Context) {
	for {
		select {
		case payload := <-ProcessPipelinesQueue:
			err := uc.execute(ctx, payload)
			if err != nil {
				uc.Log.Error("Error processing pipeline", "error", err)
			}
		case <-ctx.Done():
			uc.Log.Info("Queue processor shutting down...")
			return
		}
	}
}

func (uc *processPipelineUsecase) execute(ctx context.Context, payload external.GitHubWebhookPayload) error {
	// Only process the repository if the event is a push to the main branch
	if payload.Ref != "refs/heads/main" {
		uc.Log.Info("Ignoring event", "event", payload.Ref)
		return nil
	}

	// Extract modified files
	modifiedPipelines := payload.HeadCommit.Modified
	modifiedPipelines = util.FilterByPrefix(modifiedPipelines, "pipelines/")
	if len(modifiedPipelines) == 0 {
		log.Print("No pipeline files modified. Skipping processing.")
		return nil
	}

	// 1. Create a new branch
	newBranch := "pipeline-update-" + randomString(5)
	err := uc.GitRepository.CreateBranch(ctx, newBranch)
	if err != nil {
		return err
	}

	// 2. Switch to the new branch
	err = uc.GitRepository.SwitchBranch(ctx, newBranch)
	if err != nil {
		// Clean up in case of error
		gitCleanUp(uc, ctx, newBranch)
		uc.Log.Error("Error switching branch", "error", err)
	}

	// 3. Process pipeline files
	for _, filePath := range modifiedPipelines {
		uc.Log.Info("Initiating processing for file", "filePath", filePath)

		// Read file content
		file, err := uc.GitRepository.FindByPath(ctx, filePath)
		if err != nil {
			uc.Log.Error("FindByPath error", "filePath", filePath, "error", err)
			continue
		}

		// Pass the file content to generate the DAG
		dagContent, err := uc.GenerateAirFlowDAGUsecase.Execute(ctx, file.Content, filePath)
		if err != nil {
			uc.Log.Error("GenerateAirflowDAGUsecase error", "filePath", filePath, "error", err)
			continue
		}

		// Determine the output path (trim prefix, then change .yaml to .py)
		relativePath := strings.TrimPrefix(filePath, PIPELINES_DIRECTORY)
		dagPath := filepath.Join(OUTPUT_DIRECTORY, relativePath)
		dagPath = strings.TrimSuffix(dagPath, filepath.Ext(dagPath)) + ".py"

		uc.Log.Debug("Generated DAG content", "dagPath", dagPath)

		// Create the file in the Git repo
		generatedDAGfile := &entity.File{
			Path:    dagPath,
			Content: []byte(dagContent),
		}
		err = uc.GitRepository.Update(ctx, generatedDAGfile)
		if err != nil {
			uc.Log.Error("Error creating file", "filePath", filePath, "error", err)
			continue
		}
	}

	// 4. Commit and push changes
	commitMessage := "Automated DAG Generation"
	err = uc.GitRepository.CommitAndPush(ctx, commitMessage)
	if err != nil {
		// Clean up in case of error
		gitCleanUp(uc, ctx, newBranch)

		uc.Log.Error("Error committing and pushing changes", "error", err)
		return err
	}

	// 5. Create a pull request
	err = createPullRequest(uc, ctx, payload, newBranch)
	if err != nil {
		uc.Log.Error("Error creating pull request", "error", err)
		return err
	}

	// 6. Switch back to the main branch
	gitCleanUp(uc, ctx, newBranch)

	return nil
}

func createPullRequest(uc *processPipelineUsecase, ctx context.Context, payload external.GitHubWebhookPayload, branch string) error {
	owner := payload.Repository.Owner.Login
	repoName := payload.Repository.Name
	prTitle := "Automated DAG Generation"
	prBody := "This pull request was automatically generated to add DAGs based on pipeline definitions."
	baseBranch := uc.Config.Git.DefaultBranch
	headBranch := branch

	_, err := uc.GitHubService.CreatePullRequest(
		owner,
		repoName,
		prTitle,
		headBranch,
		baseBranch,
		prBody,
		uc.Config.Git.Token,
	)
	if err != nil {
		// Clean up in case of error
		gitCleanUp(uc, ctx, branch)
		return err
	}

	return nil
}

func gitCleanUp(uc *processPipelineUsecase, ctx context.Context, branchName string) error {
	err := uc.GitRepository.SwitchBackToMain(ctx)
	if err != nil {
		return err
	}

	err = uc.GitRepository.DeleteBranch(ctx, branchName)
	if err != nil {
		return err
	}

	return nil
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomString(length int) string {
	bytes := make([]byte, length)
	for i := range bytes {
		randomByte, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		bytes[i] = charset[randomByte.Int64()]
	}
	return string(bytes)
}
