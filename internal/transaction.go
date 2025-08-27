package internal

import "context"

type TransactionManager interface {
	Do(ctx context.Context, branch, message string, fn func(ctx context.Context) error) error
	Checkout(ctx context.Context, branch string, fn func(ctx context.Context) error) error
	GetMergeBase(ctx context.Context, source, branch string) (string, error)
	GetHead(ctx context.Context) (string, error)
	MergeBranch(ctx context.Context, branch string) error
	DeleteBranch(ctx context.Context, branch string) error
	BranchExists(ctx context.Context, branch string) (bool, error)
}
