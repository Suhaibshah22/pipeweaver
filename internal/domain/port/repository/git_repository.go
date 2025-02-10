package repository

import (
	"context"

	"github.com/Suhaibshah22/pipeweaver/internal/domain/entity"
)

type GitRepository interface {
	FindByPath(ctx context.Context, path string) (*entity.File, error)

	CommitAndPush(ctx context.Context, message string) error

	CreateBranch(ctx context.Context, branchName string) error

	SwitchBranch(ctx context.Context, branchName string) error

	Create(ctx context.Context, file *entity.File) error

	Update(ctx context.Context, file *entity.File) error

	SwitchBackToMain(ctx context.Context) error

	DeleteBranch(ctx context.Context, branchName string) error
}
