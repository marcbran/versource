package merge

import (
	"context"
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal/tui/platform"
	"github.com/marcbran/versource/pkg/versource"
)

type DetailData struct {
	facade        versource.Facade
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

func NewDetail(facade versource.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewViewDataViewport(NewDetailData(facade, params["changesetName"], params["mergeID"]))
	}
}

func NewDetailData(facade versource.Facade, changesetName, mergeID string) *DetailData {
	return &DetailData{
		facade:        facade,
		changesetName: changesetName,
		mergeID:       mergeID,
	}
}

func (p *DetailData) LoadData() (*versource.GetMergeResponse, error) {
	ctx := context.Background()
	mergeID, err := strconv.ParseUint(p.mergeID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid merge ID: %w", err)
	}

	req := versource.GetMergeRequest{
		MergeID:       uint(mergeID),
		ChangesetName: p.changesetName,
	}

	resp, err := p.facade.GetMerge(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (p *DetailData) ResolveData(data versource.GetMergeResponse) DetailViewModel {
	return DetailViewModel{
		ID:          data.Merge.ID,
		ChangesetID: data.Merge.ChangesetID,
		State:       string(data.Merge.State),
		MergeBase:   data.Merge.MergeBase,
		Head:        data.Merge.Head,
	}
}

func (p *DetailData) KeyBindings(elem versource.GetMergeResponse) platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "esc", Help: "View merges", Command: fmt.Sprintf("changesets/%s/merges", p.changesetName)},
	}
}
