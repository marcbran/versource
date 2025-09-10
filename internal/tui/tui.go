package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui/apply"
	"github.com/marcbran/versource/internal/tui/changeset"
	"github.com/marcbran/versource/internal/tui/component"
	"github.com/marcbran/versource/internal/tui/module"
	"github.com/marcbran/versource/internal/tui/plan"
	"github.com/marcbran/versource/internal/tui/platform"
)

func RunApp(client *client.Client) error {
	router := platform.NewRouter().
		KeyBinding("", func(params map[string]string) platform.KeyBindings {
			return platform.KeyBindings{
				{Key: "m", Help: "View modules", Command: "modules"},
				{Key: "g", Help: "View changesets", Command: "changesets"},
				{Key: "c", Help: "View components", Command: "components"},
				{Key: "p", Help: "View plans", Command: "plans"},
				{Key: "a", Help: "View applies", Command: "applies"},
			}
		}).
		KeyBinding("changesets/{changesetName}", func(params map[string]string) platform.KeyBindings {
			changesetName := params["changesetName"]
			return platform.KeyBindings{
				{Key: "m", Help: "View modules", Command: fmt.Sprintf("changesets/%s/modules", changesetName)},
				{Key: "c", Help: "View components", Command: fmt.Sprintf("changesets/%s/components", changesetName)},
				{Key: "d", Help: "View component diffs", Command: fmt.Sprintf("changesets/%s/components/diffs", changesetName)},
				{Key: "p", Help: "View plans", Command: fmt.Sprintf("changesets/%s/plans", changesetName)},
				{Key: "a", Help: "View applies", Command: fmt.Sprintf("changesets/%s/applies", changesetName)},
			}
		}).
		Route("modules", module.NewTable(client)).
		Route("modules/{moduleID}", module.NewDetail(client)).
		Route("modules/{moduleID}/moduleversions", module.NewVersionsForModuleTable(client)).
		Route("moduleversions", module.NewVersionsTable(client)).
		Route("moduleversions/{moduleVersionID}", module.NewVersionDetail(client)).
		Route("components", component.NewTable(client)).
		Route("components/{componentID}", component.NewDetail(client)).
		Route("plans", plan.NewTable(client)).
		Route("plans/{planID}/logs", plan.NewLogs(client)).
		Route("applies", apply.NewTable(client)).
		Route("changesets", changeset.NewTable(client)).
		Route("changesets/{changesetName}/components", component.NewChangesetTable(client)).
		Route("changesets/{changesetName}/components/diffs", component.NewChangesetDiffTable(client)).
		Route("changesets/{changesetName}/plans", plan.NewChangesetTable(client))

	app := platform.NewCommandable(router, client)

	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to run: %w", err)
	}

	return nil
}
