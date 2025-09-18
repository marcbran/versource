package merge

import (
	"context"
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/tui/platform"
	"gopkg.in/yaml.v3"
)

type DetailData struct {
	facade        internal.Facade
	changesetName string
	mergeID       string
}

type DetailViewModel struct {
	ID          uint   `yaml:"id"`
	ChangesetID uint   `yaml:"changeset_id"`
	State       string `yaml:"state"`
	MergeBase   string `yaml:"merge_base"`
	Head        string `yaml:"head"`
}

func NewDetail(facade internal.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewDataViewport(NewDetailData(facade, params["changesetName"], params["mergeID"]))
	}
}

func NewDetailData(facade internal.Facade, changesetName, mergeID string) *DetailData {
	return &DetailData{
		facade:        facade,
		changesetName: changesetName,
		mergeID:       mergeID,
	}
}

func (p *DetailData) LoadData() (*internal.GetMergeResponse, error) {
	ctx := context.Background()
	mergeID, err := strconv.ParseUint(p.mergeID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid merge ID: %w", err)
	}

	req := internal.GetMergeRequest{
		MergeID:       uint(mergeID),
		ChangesetName: p.changesetName,
	}

	resp, err := p.facade.GetMerge(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (p *DetailData) ResolveData(data internal.GetMergeResponse) string {
	viewModel := DetailViewModel{
		ID:          data.ID,
		ChangesetID: data.ChangesetID,
		State:       string(data.State),
		MergeBase:   data.MergeBase,
		Head:        data.Head,
	}

	yamlData, err := yaml.Marshal(viewModel)
	if err != nil {
		return fmt.Sprintf("Error marshaling data: %v", err)
	}

	return string(yamlData)
}

func (p *DetailData) KeyBindings(elem internal.GetMergeResponse) platform.KeyBindings {
	return platform.KeyBindings{}
}
