package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/tui/apply"
	"github.com/marcbran/versource/internal/tui/changeset"
	"github.com/marcbran/versource/internal/tui/component"
	"github.com/marcbran/versource/internal/tui/module"
	"github.com/marcbran/versource/internal/tui/plan"
	"github.com/marcbran/versource/internal/tui/platform"
)

func RunApp(facade internal.Facade) error {
	router := platform.NewRouter("components").
		KeyBinding("", func(params map[string]string, currentPath string) platform.KeyBindings {
			return platform.KeyBindings{
				{Key: "r", Help: "Refresh", Command: "refresh"},
				{Key: "b", Help: "Go back", Command: "back"},
				{Key: "m", Help: "View modules", Command: "modules"},
				{Key: "g", Help: "View changesets", Command: "changesets"},
				{Key: "c", Help: "View components", Command: "components"},
				{Key: "p", Help: "View plans", Command: "plans"},
				{Key: "a", Help: "View applies", Command: "applies"},
			}
		}).
		KeyBinding("changesets/{changesetName}", func(params map[string]string, currentPath string) platform.KeyBindings {
			changesetName := params["changesetName"]
			pathWithoutChangeset := removeFirstTwoSegments(currentPath)
			return platform.KeyBindings{
				{Key: "esc", Help: "Back to changesets", Command: pathWithoutChangeset},
				{Key: "m", Help: "View modules", Command: fmt.Sprintf("changesets/%s/modules", changesetName)},
				{Key: "c", Help: "View components", Command: fmt.Sprintf("changesets/%s/components", changesetName)},
				{Key: "d", Help: "View diffs", Command: fmt.Sprintf("changesets/%s/diffs", changesetName)},
				{Key: "p", Help: "View plans", Command: fmt.Sprintf("changesets/%s/plans", changesetName)},
			}
		}).
		Route("modules", module.NewTable(facade)).
		Route("modules/create", module.NewCreateModule(facade)).
		Route("modules/{moduleID}", module.NewDetail(facade)).
		Route("modules/{moduleID}/delete", module.NewDeleteModule(facade)).
		Route("modules/{moduleID}/moduleversions", module.NewVersionsTable(facade)).
		Route("moduleversions", module.NewVersionsTable(facade)).
		Route("moduleversions/{moduleVersionID}", module.NewVersionDetail(facade)).
		Route("components", component.NewTable(facade)).
		Route("components/create", component.NewCreateComponent(facade)).
		Route("components/{componentID}", component.NewDetail(facade)).
		Route("components/{componentID}/delete", component.NewDelete(facade)).
		Route("plans", plan.NewTable(facade)).
		Route("plans/{planID}/logs", plan.NewLogs(facade)).
		Route("applies", apply.NewTable(facade)).
		Route("changesets", changeset.NewTable(facade)).
		Route("changesets/{changesetName}/components", component.NewChangesetTable(facade)).
		Route("changesets/{changesetName}/diffs", component.NewChangesetDiffsTable(facade)).
		Route("changesets/{changesetName}/components/{componentID}/plans/create", component.NewCreatePlan(facade)).
		Route("changesets/{changesetName}/components/{componentID}/delete", component.NewDelete(facade)).
		Route("changesets/{changesetName}/components/{componentID}/restore", component.NewRestore(facade)).
		Route("changesets/{changesetName}/plans", plan.NewChangesetTable(facade))

	app := platform.NewCommandable(router, facade)

	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}

	return nil
}

func removeFirstTwoSegments(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) <= 2 {
		return ""
	}
	return strings.Join(parts[2:], "/")
}
