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
	GetBranchHead(ctx context.Context, branch string) (string, error)
	HasCommitsAfter(ctx context.Context, branch, commit string) (bool, error)
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

func IsValidBranch(branch string) bool {
	if branch == "" {
		return false
	}

	if len(branch) == 32 && IsValidCommitHash(branch) {
		return false
	}

	if branch[0] == '.' {
		return false
	}

	if len(branch) >= 2 {
		for i := 0; i < len(branch)-1; i++ {
			if branch[i] == '.' && branch[i+1] == '.' {
				return false
			}
		}
	}

	for i := 0; i < len(branch)-1; i++ {
		if branch[i] == '@' && branch[i+1] == '{' {
			return false
		}
		if branch[i] == '-' && branch[i+1] == '-' {
			return false
		}
		if branch[i] == '/' && branch[i+1] == '*' {
			return false
		}
		if branch[i] == '*' && branch[i+1] == '/' {
			return false
		}
	}

	for _, char := range branch {
		if char < 32 || char > 126 {
			return false
		}

		switch char {
		case ':', '?', '[', '\\', '^', '~', '*':
			return false
		case ' ', '\t', '\n', '\r':
			return false
		case '\'', '"', '`', ';', '(', ')', '=', '<', '>', '|', '&', '$', '%', '+', '!', '#':
			return false
		}
	}

	if branch[len(branch)-1] == '/' {
		return false
	}

	if len(branch) >= 5 && branch[len(branch)-5:] == ".lock" {
		return false
	}

	if len(branch) == 4 && (branch[0] == 'H' || branch[0] == 'h') && (branch[1] == 'E' || branch[1] == 'e') && (branch[2] == 'A' || branch[2] == 'a') && (branch[3] == 'D' || branch[3] == 'd') {
		return false
	}

	return true
}
