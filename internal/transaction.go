package internal

import "context"

const MainBranch = "main"

type TransactionManager interface {
	Do(ctx context.Context, branch, message string, fn func(ctx context.Context) error) error
	Checkout(ctx context.Context, branch string, fn func(ctx context.Context) error) error

	HasBranch(ctx context.Context, branch string) (bool, error)
	CreateBranch(ctx context.Context, branch string) error
	MergeBranch(ctx context.Context, branch string) error
	DeleteBranch(ctx context.Context, branch string) error

	GetMergeBase(ctx context.Context, source, branch string) (string, error)
	GetHead(ctx context.Context) (string, error)
}

func IsValidCommitHash(hash string) bool {
	if len(hash) != 32 {
		return false
	}

	for _, char := range hash {
		if (char < '0' || char > '9') && (char < 'a' || char > 'z') {
			return false
		}
	}

	return true
}
