package plan

import (
	"context"
	"io"
	"strconv"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/tui/platform"
)

type LogsData struct {
	facade internal.Facade
	planID string
}

func NewLogs(facade internal.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataViewport(&LogsData{facade: facade, planID: params["planID"]})
	}
}

func (p *LogsData) LoadData() (*internal.GetPlanLogResponse, error) {
	ctx := context.Background()

	planIDUint, err := strconv.ParseUint(p.planID, 10, 32)
	if err != nil {
		return nil, err
	}

	resp, err := p.facade.GetPlanLog(ctx, internal.GetPlanLogRequest{PlanID: uint(planIDUint)})
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
	return platform.KeyBindings{}
}
