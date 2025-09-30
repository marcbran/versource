package plan

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/tui/platform"
)

type LogsData struct {
	facade        internal.Facade
	changesetName string
	planID        string
}

func NewLogs(facade internal.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataViewport(NewLogsData(
			facade,
			params["changesetName"],
			params["planID"],
		))
	}
}

func NewLogsData(facade internal.Facade, changesetName string, planID string) *LogsData {
	return &LogsData{
		facade:        facade,
		changesetName: changesetName,
		planID:        planID,
	}
}

func (p *LogsData) LoadData() (*internal.GetPlanLogResponse, error) {
	ctx := context.Background()

	planIDUint, err := strconv.ParseUint(p.planID, 10, 32)
	if err != nil {
		return nil, err
	}

	req := internal.GetPlanLogRequest{PlanID: uint(planIDUint)}
	if p.changesetName != "" {
		req.ChangesetName = &p.changesetName
	}

	resp, err := p.facade.GetPlanLog(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *LogsData) ResolveData(data internal.GetPlanLogResponse) string {
	content, err := io.ReadAll(data.Content)
	if err != nil {
		return "Failed to read log content"
	}
	defer data.Content.Close()

	return string(content)
}

func (p *LogsData) KeyBindings(elem internal.GetPlanLogResponse) platform.KeyBindings {
	changesetPrefix := ""
	if p.changesetName != "" {
		changesetPrefix = fmt.Sprintf("changesets/%s", p.changesetName)
	}
	return platform.KeyBindings{
		{Key: "esc", Help: "View plan", Command: fmt.Sprintf("%s/plans/%s", changesetPrefix, p.planID)},
	}
}
