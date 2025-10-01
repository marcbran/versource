package plan

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
	planID        string
}

type DetailViewModel struct {
	ID        uint   `yaml:"id"`
	State     string `yaml:"state"`
	From      string `yaml:"from"`
	To        string `yaml:"to"`
	Add       *int   `yaml:"add,omitempty"`
	Change    *int   `yaml:"change,omitempty"`
	Destroy   *int   `yaml:"destroy,omitempty"`
	Component *struct {
		ID   uint   `yaml:"id"`
		Name string `yaml:"name"`
	} `yaml:"component,omitempty"`
	Changeset *struct {
		ID   uint   `yaml:"id"`
		Name string `yaml:"name"`
	} `yaml:"changeset,omitempty"`
}

func NewDetail(facade internal.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewViewDataViewport(NewDetailData(
			facade,
			params["changesetName"],
			params["planID"],
		))
	}
}

func NewDetailData(facade internal.Facade, changesetName string, planID string) *DetailData {
	return &DetailData{
		facade:        facade,
		changesetName: changesetName,
		planID:        planID,
	}
}

func (p *DetailData) LoadData() (*internal.GetPlanResponse, error) {
	ctx := context.Background()

	planIDUint, err := strconv.ParseUint(p.planID, 10, 32)
	if err != nil {
		return nil, err
	}

	req := internal.GetPlanRequest{PlanID: uint(planIDUint)}
	if p.changesetName != "" {
		req.ChangesetName = &p.changesetName
	}

	planResp, err := p.facade.GetPlan(ctx, req)
	if err != nil {
		return nil, err
	}

	return planResp, nil
}

func (p *DetailData) ResolveData(data internal.GetPlanResponse) DetailViewModel {
	var component *struct {
		ID   uint   `yaml:"id"`
		Name string `yaml:"name"`
	}
	if data.Component.ID != 0 {
		component = &struct {
			ID   uint   `yaml:"id"`
			Name string `yaml:"name"`
		}{
			ID:   data.Component.ID,
			Name: data.Component.Name,
		}
	}

	var changeset *struct {
		ID   uint   `yaml:"id"`
		Name string `yaml:"name"`
	}
	if data.Changeset.ID != 0 {
		changeset = &struct {
			ID   uint   `yaml:"id"`
			Name string `yaml:"name"`
		}{
			ID:   data.Changeset.ID,
			Name: data.Changeset.Name,
		}
	}

	return DetailViewModel{
		ID:        data.ID,
		State:     string(data.State),
		From:      data.From,
		To:        data.To,
		Add:       data.Add,
		Change:    data.Change,
		Destroy:   data.Destroy,
		Component: component,
		Changeset: changeset,
	}
}

func (p *DetailData) KeyBindings(elem internal.GetPlanResponse) platform.KeyBindings {
	changesetPrefix := ""
	if p.changesetName != "" {
		changesetPrefix = fmt.Sprintf("changesets/%s", p.changesetName)
	}
	return platform.KeyBindings{
		{Key: "esc", Help: "View plans", Command: fmt.Sprintf("%s/plans", changesetPrefix)},
		{Key: "l", Help: "View logs", Command: fmt.Sprintf("%s/plans/%s/logs", changesetPrefix, p.planID)},
		{Key: "c", Help: "View component", Command: fmt.Sprintf("%s/components/%d", changesetPrefix, elem.ComponentID)},
	}
}
