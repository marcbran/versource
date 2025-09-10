package plan

import (
	"context"
	"io"
	"strconv"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/http/client"
	"github.com/marcbran/versource/internal/tui/platform"
)

type LogsData struct {
	client *client.Client
	planID string
}

func NewLogs(client *client.Client) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataViewport(&LogsData{client: client, planID: params["planID"]})
	}
}

func (p *LogsData) LoadData() (*internal.GetPlanLogResponse, error) {
	ctx := context.Background()

	planIDUint, err := strconv.ParseUint(p.planID, 10, 32)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.GetPlanLog(ctx, uint(planIDUint))
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
