package plan

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/marcbran/versource/internal/tui/platform"
	"github.com/marcbran/versource/pkg/versource"
)

type LogsData struct {
	facade        versource.Facade
	changesetName string
	planID        string
}

func NewLogs(facade versource.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataViewport(NewLogsData(
			facade,
			params["changesetName"],
			params["planID"],
		))
	}
}

func NewLogsData(facade versource.Facade, changesetName string, planID string) *LogsData {
	return &LogsData{
		facade:        facade,
		changesetName: changesetName,
		planID:        planID,
	}
}

func (p *LogsData) LoadData() (*versource.GetPlanLogResponse, error) {
	ctx := context.Background()

	planIDUint, err := strconv.ParseUint(p.planID, 10, 32)
	if err != nil {
		return nil, err
	}

	req := versource.GetPlanLogRequest{PlanID: uint(planIDUint)}
	if p.changesetName != "" {
		req.ChangesetName = &p.changesetName
	}

	resp, err := p.facade.GetPlanLog(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *LogsData) ResolveData(data versource.GetPlanLogResponse) string {
	content, err := io.ReadAll(data.Content)
	if err != nil {
		return "Failed to read log content"
	}
	defer data.Content.Close()

	return string(content)
}

func (p *LogsData) KeyBindings(elem versource.GetPlanLogResponse) platform.KeyBindings {
	changesetPrefix := ""
	if p.changesetName != "" {
		changesetPrefix = fmt.Sprintf("changesets/%s", p.changesetName)
	}
	return platform.KeyBindings{
		{Key: "esc", Help: "View plan", Command: fmt.Sprintf("%s/plans/%s", changesetPrefix, p.planID)},
	}
}
