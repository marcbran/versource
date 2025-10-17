package plan

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

func NewDetail(facade versource.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewViewDataViewport(NewDetailData(
			facade,
			params["changesetName"],
			params["planID"],
		))
	}
}

func NewDetailData(facade versource.Facade, changesetName string, planID string) *DetailData {
	return &DetailData{
		facade:        facade,
		changesetName: changesetName,
		planID:        planID,
	}
}

func (p *DetailData) LoadData() (*versource.GetPlanResponse, error) {
	ctx := context.Background()

	planIDUint, err := strconv.ParseUint(p.planID, 10, 32)
	if err != nil {
		return nil, err
	}

	req := versource.GetPlanRequest{PlanID: uint(planIDUint)}
	if p.changesetName != "" {
		req.ChangesetName = &p.changesetName
	}

	planResp, err := p.facade.GetPlan(ctx, req)
	if err != nil {
		return nil, err
	}

	return planResp, nil
}

func (p *DetailData) ResolveData(data versource.GetPlanResponse) DetailViewModel {
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
	if data.Plan.Changeset.ID != 0 {
		changeset = &struct {
			ID   uint   `yaml:"id"`
			Name string `yaml:"name"`
		}{
			ID:   data.Plan.Changeset.ID,
			Name: data.Plan.Changeset.Name,
		}
	}

	return DetailViewModel{
		ID:        data.Plan.ID,
		State:     string(data.Plan.State),
		From:      data.Plan.From,
		To:        data.Plan.To,
		Add:       data.Plan.Add,
		Change:    data.Plan.Change,
		Destroy:   data.Plan.Destroy,
		Component: component,
		Changeset: changeset,
	}
}

func (p *DetailData) KeyBindings(elem versource.GetPlanResponse) platform.KeyBindings {
	changesetPrefix := ""
	if p.changesetName != "" {
		changesetPrefix = fmt.Sprintf("changesets/%s", p.changesetName)
	}
	return platform.KeyBindings{
		{Key: "esc", Help: "View plans", Command: fmt.Sprintf("%s/plans", changesetPrefix)},
		{Key: "l", Help: "View logs", Command: fmt.Sprintf("%s/plans/%s/logs", changesetPrefix, p.planID)},
		{Key: "c", Help: "View component", Command: fmt.Sprintf("%s/components/%d", changesetPrefix, elem.Plan.ComponentID)},
	}
}
