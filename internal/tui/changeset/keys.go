package changeset

import (
	"fmt"

	"github.com/marcbran/versource/internal/tui/platform"
)

func KeyBindings(changesetName string) platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "m", Help: "View modules", Command: fmt.Sprintf("changesets/%s/modules", changesetName)},
		{Key: "c", Help: "View components", Command: fmt.Sprintf("changesets/%s/components", changesetName)},
		{Key: "d", Help: "View component diffs", Command: fmt.Sprintf("changesets/%s/components/diffs", changesetName)},
		{Key: "p", Help: "View plans", Command: fmt.Sprintf("changesets/%s/plans", changesetName)},
		{Key: "a", Help: "View applies", Command: fmt.Sprintf("changesets/%s/applies", changesetName)},
	}
}
