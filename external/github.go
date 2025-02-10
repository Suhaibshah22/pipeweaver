package external

import (
	"context"

	"github.com/google/go-github/v50/github"
	"golang.org/x/oauth2"
)

type GitHubService interface {
	CreatePullRequest(owner, repo, title, head, base, body, token string) (*github.PullRequest, error)
}

type gitHubService struct{}

func NewGitHubService() GitHubService {
	return &gitHubService{}
}

func (p *gitHubService) CreatePullRequest(owner, repo, title, head, base, body, token string) (*github.PullRequest, error) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	newPR := &github.NewPullRequest{
		Title: github.String(title),
		Head:  github.String(head),
		Base:  github.String(base),
		Body:  github.String(body),
	}

	pr, _, err := client.PullRequests.Create(ctx, owner, repo, newPR)
	if err != nil {
		return nil, err
	}

	return pr, nil
}

type GitHubWebhookPayload struct {
	Ref        string `json:"ref"`
	Before     string `json:"before"`
	After      string `json:"after"`
	Repository struct {
		Name  string `json:"name"`
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
		CloneURL string `json:"clone_url"`
	} `json:"repository"`
	Commits []struct {
		ID       string   `json:"id"`
		Message  string   `json:"message"`
		Modified []string `json:"modified"`
	} `json:"commits"`
	HeadCommit struct {
		ID       string   `json:"id"`
		Message  string   `json:"message"`
		Modified []string `json:"modified"`
	} `json:"head_commit"`
}
