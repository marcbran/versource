package rebase

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
	rebaseID      string
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
		return platform.NewDataViewport(NewDetailData(facade, params["changesetName"], params["rebaseID"]))
	}
}

func NewDetailData(facade internal.Facade, changesetName, rebaseID string) *DetailData {
	return &DetailData{
		facade:        facade,
		changesetName: changesetName,
		rebaseID:      rebaseID,
	}
}

func (p *DetailData) LoadData() (*internal.GetRebaseResponse, error) {
	ctx := context.Background()
	rebaseID, err := strconv.ParseUint(p.rebaseID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid rebase ID: %w", err)
	}

	req := internal.GetRebaseRequest{
		RebaseID:      uint(rebaseID),
		ChangesetName: p.changesetName,
	}

	resp, err := p.facade.GetRebase(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (p *DetailData) ResolveData(data internal.GetRebaseResponse) string {
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

func (p *DetailData) KeyBindings(elem internal.GetRebaseResponse) platform.KeyBindings {
	return platform.KeyBindings{}
}
