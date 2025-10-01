package merge

import (
	"context"
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/tui/platform"
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
		return platform.NewViewDataViewport(NewDetailData(facade, params["changesetName"], params["mergeID"]))
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

func (p *DetailData) ResolveData(data internal.GetMergeResponse) DetailViewModel {
	return DetailViewModel{
		ID:          data.ID,
		ChangesetID: data.ChangesetID,
		State:       string(data.State),
		MergeBase:   data.MergeBase,
		Head:        data.Head,
	}
}

func (p *DetailData) KeyBindings(elem internal.GetMergeResponse) platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "esc", Help: "View merges", Command: fmt.Sprintf("changesets/%s/merges", p.changesetName)},
	}
}
