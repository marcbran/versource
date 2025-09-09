package tui

import (
	"github.com/marcbran/versource/internal/tui/platform"
)

var KeyBindings = platform.KeyBindings{
	{Key: "m", Help: "View modules", Command: "modules"},
	{Key: "g", Help: "View changesets", Command: "changesets"},
	{Key: "c", Help: "View components", Command: "components"},
	{Key: "p", Help: "View plans", Command: "plans"},
	{Key: "a", Help: "View applies", Command: "applies"},
}
