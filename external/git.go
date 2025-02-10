package external

/*
This file will implement all of the low level git operations,
so that the repositories that treat this like any other repository driver.
*/

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

type GitService interface {
	CloneRepo(repoURL, branch, directory, username, token string) (*git.Repository, error)
	PullRepo(repo *git.Repository, branch string) error
}

type gitService struct{}

func NewGitService() GitService {
	return &gitService{}
}

func (g *gitService) CloneRepo(repoURL, branch, directory, username, token string) (*git.Repository, error) {
	// Check if directory exists
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		// Clone the repository
		repo, err := git.PlainClone(directory, false, &git.CloneOptions{
			URL:           repoURL,
			ReferenceName: plumbing.ReferenceName(branch),
			SingleBranch:  true,
			Depth:         1,
			Auth: &http.BasicAuth{
				Username: username, // can be anything except an empty string
				Password: token,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to clone repository: %w", err)
		}
		return repo, nil
	} else {
		// Open the existing repository
		repo, err := git.PlainOpen(directory)
		if err != nil {
			return nil, fmt.Errorf("failed to open repository: %w", err)
		}
		// Pull the latest changes
		err = g.PullRepo(repo, branch)
		if err != nil && err != git.NoErrAlreadyUpToDate {
			return nil, fmt.Errorf("failed to pull repository: %w", err)
		}
		return repo, nil
	}
}

func (g *gitService) PullRepo(repo *git.Repository, branch string) error {
	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}

	err = w.Pull(&git.PullOptions{
		ReferenceName: plumbing.ReferenceName(branch),
		SingleBranch:  true,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return fmt.Errorf("failed to pull repository: %w", err)
	}

	return nil
}

// Utility function to get absolute path
func GetRepoDirectory(baseDir, repoName string) string {
	return filepath.Join(baseDir, repoName)
}
