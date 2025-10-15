package apply

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/marcbran/versource/internal/tui/platform"
	"github.com/marcbran/versource/pkg/versource"
)

type LogsData struct {
	facade  versource.Facade
	applyID string
}

func NewLogs(facade versource.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataViewport(NewLogsData(
			facade,
			params["applyID"],
		))
	}
}

func NewLogsData(facade versource.Facade, applyID string) *LogsData {
	return &LogsData{
		facade:  facade,
		applyID: applyID,
	}
}

func (p *LogsData) LoadData() (*versource.GetApplyLogResponse, error) {
	ctx := context.Background()

	applyIDUint, err := strconv.ParseUint(p.applyID, 10, 32)
	if err != nil {
		return nil, err
	}

	req := versource.GetApplyLogRequest{ApplyID: uint(applyIDUint)}

	resp, err := p.facade.GetApplyLog(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (p *LogsData) ResolveData(data versource.GetApplyLogResponse) string {
	content, err := io.ReadAll(data.Content)
	if err != nil {
		return "Failed to read log content"
	}
	defer data.Content.Close()

	return string(content)
}

func (p *LogsData) KeyBindings(elem versource.GetApplyLogResponse) platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "esc", Help: "View apply", Command: fmt.Sprintf("applies/%s", p.applyID)},
	}
}
