package repository

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Suhaibshah22/pipeweaver/internal/domain/entity"
	"github.com/Suhaibshah22/pipeweaver/internal/domain/port/repository"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type gitRepositoryImpl struct {
	Repo     *git.Repository
	Worktree *git.Worktree
	Auth     *http.BasicAuth

	RepoPath  string
	RemoteURL string
}

func NewGitRepository(remoteURL, branch, repoPath, username, token string) (repository.GitRepository, error) {
	var repo *git.Repository
	var err error

	auth := &http.BasicAuth{
		Username: username,
		Password: token,
	}

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		log.Printf("Cloning repository from %s into %s", remoteURL, repoPath)
		//Clone the repository
		repo, err = git.PlainClone(repoPath, false, &git.CloneOptions{
			URL:           remoteURL,
			ReferenceName: plumbing.NewBranchReferenceName(branch),
			SingleBranch:  true,
			// Depth:         1,     //Potentially causing a null object and not pulling the latest changes
			Auth: auth,
		})
		if err != nil {
			log.Printf("Clone failed: %v", err)
			return nil, fmt.Errorf("failed to clone repository: %w", err)
		}
		log.Println("Repository cloned successfully.")
	} else {
		log.Printf("Opening existing repository at %s", repoPath)
		// Open the existing repository
		repo, err = git.PlainOpen(repoPath)
		if err != nil {
			log.Printf("Pull failed: %v", err)
			return nil, fmt.Errorf("failed to open repository: %w", err)
		}

		// Pull the latest changes
		worktree, err := repo.Worktree()
		if err != nil {
			return nil, fmt.Errorf("failed to get worktree: %w", err)
		}

		err = worktree.Pull(&git.PullOptions{
			ReferenceName: plumbing.NewBranchReferenceName(branch),
			SingleBranch:  true,
			Auth:          auth,
		})
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return nil, fmt.Errorf("failed to pull repository: %w", err)
		}
		log.Println("Repository pulled successfully.")
	}

	// Get the worktree
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	return &gitRepositoryImpl{
		Repo:      repo,
		Worktree:  worktree,
		Auth:      auth,
		RepoPath:  repoPath,
		RemoteURL: remoteURL,
	}, nil

}

// FindByPath implements repository.GitRepository.
func (g *gitRepositoryImpl) FindByPath(ctx context.Context, path string) (*entity.File, error) {
	fullPath := filepath.Join(g.RepoPath, path)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return &entity.File{
		Path:    path,
		Content: content,
	}, nil
}

// CreateBranch implements repository.GitRepository.
func (g *gitRepositoryImpl) CreateBranch(ctx context.Context, branchName string) error {
	headRef, err := g.Repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD reference: %w", err)
	}

	newBranchRef := plumbing.NewHashReference(plumbing.NewBranchReferenceName(branchName), headRef.Hash())
	err = g.Repo.Storer.SetReference(newBranchRef)
	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	return g.SwitchBranch(ctx, branchName)
}

// SwitchBranch implements repository.GitRepository.
func (g *gitRepositoryImpl) SwitchBranch(ctx context.Context, branchName string) error {
	err := g.Worktree.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName(branchName)})
	if err != nil {
		return fmt.Errorf("failed to switch branch: %w", err)
	}
	return nil
}

// CommitAndPush implements repository.GitRepository.
func (g *gitRepositoryImpl) CommitAndPush(ctx context.Context, message string) error {
	// Commit the changes
	_, err := g.Worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Go Clean Arch Bot", // Change this to be supplied via configs
			Email: "gitops@email.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to commit changes: %w", err)
	}

	//Push the changes
	err = g.Repo.Push(&git.PushOptions{
		Auth: g.Auth,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to push changes: %w", err)
	}

	return nil
}

// Update implements repository.GitRepository.
func (g *gitRepositoryImpl) Update(ctx context.Context, file *entity.File) error {
	// 1) Join repo path + file path
	fullPath := filepath.Join(g.RepoPath, file.Path)

	// 2) Ensure directory exists at the same (joined) path
	if err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", filepath.Dir(fullPath), err)
	}

	log.Printf("Creating file at path: %s", fullPath)

	// 3) Write the file
	err := os.WriteFile(fullPath, file.Content, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file at %s: %w", fullPath, err)
	}

	// 4) Add file to Git worktree
	_, err = g.Worktree.Add(file.Path)
	if err != nil {
		return fmt.Errorf("failed to add file %s to worktree: %w", file.Path, err)
	}

	log.Printf("Successfully created and added file: %s", file.Path)
	return nil
}

// Create implements repository.GitRepository.
func (g *gitRepositoryImpl) Create(ctx context.Context, file *entity.File) error {
	return g.Update(ctx, file) // Similar to Update
}

// SwitchBackToMain implements repository.GitRepository.
func (g *gitRepositoryImpl) SwitchBackToMain(ctx context.Context) error {
	mainBranch := "main"

	// log.Print("Switching back to main branch", "branchName", mainBranch)

	// Checkout the main branch
	err := g.Worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(mainBranch),
	})
	if err != nil {
		// g.Logger.Error("Failed to switch back to main branch", "branchName", mainBranch, "error", err)
		return fmt.Errorf("failed to switch back to main branch: %w", err)
	}

	// g.Logger.Info("Switched back to main branch successfully", "branchName", mainBranch)
	return nil
}

// DeleteBranch implements repository.GitRepository.
func (g *gitRepositoryImpl) DeleteBranch(ctx context.Context, branchName string) error {

	err := g.Repo.DeleteBranch(branchName)
	if err != nil {
		return fmt.Errorf("failed to delete branch: %w", err)
	}
	return nil
}
