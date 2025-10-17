package rebase

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
	rebaseID      string
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
		return platform.NewViewDataViewport(NewDetailData(facade, params["changesetName"], params["rebaseID"]))
	}
}

func NewDetailData(facade versource.Facade, changesetName, rebaseID string) *DetailData {
	return &DetailData{
		facade:        facade,
		changesetName: changesetName,
		rebaseID:      rebaseID,
	}
}

func (p *DetailData) LoadData() (*versource.GetRebaseResponse, error) {
	ctx := context.Background()
	rebaseID, err := strconv.ParseUint(p.rebaseID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid rebase ID: %w", err)
	}

	req := versource.GetRebaseRequest{
		RebaseID:      uint(rebaseID),
		ChangesetName: p.changesetName,
	}

	resp, err := p.facade.GetRebase(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (p *DetailData) ResolveData(data versource.GetRebaseResponse) DetailViewModel {
	return DetailViewModel{
		ID:          data.Rebase.ID,
		ChangesetID: data.Rebase.ChangesetID,
		State:       string(data.Rebase.State),
		MergeBase:   data.Rebase.MergeBase,
		Head:        data.Rebase.Head,
	}
}

func (p *DetailData) KeyBindings(elem versource.GetRebaseResponse) platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "esc", Help: "View rebases", Command: fmt.Sprintf("changesets/%s/rebases", p.changesetName)},
	}
}
