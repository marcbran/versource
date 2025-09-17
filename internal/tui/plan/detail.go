package plan

import (
	"context"
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/tui/platform"
	"gopkg.in/yaml.v3"
)

type DetailData struct {
	facade internal.Facade
	planID string
}

type DetailViewModel struct {
	ID        uint   `yaml:"id"`
	State     string `yaml:"state"`
	MergeBase string `yaml:"merge_base"`
	Head      string `yaml:"head"`
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
		return platform.NewDataViewport(NewDetailData(
			facade,
			params["planID"],
		))
	}
}

func NewDetailData(facade internal.Facade, planID string) *DetailData {
	return &DetailData{
		facade: facade,
		planID: planID,
	}
}

func (p *DetailData) LoadData() (*internal.GetPlanResponse, error) {
	ctx := context.Background()

	planIDUint, err := strconv.ParseUint(p.planID, 10, 32)
	if err != nil {
		return nil, err
	}

	planResp, err := p.facade.GetPlan(ctx, internal.GetPlanRequest{PlanID: uint(planIDUint)})
	if err != nil {
		return nil, err
	}

	return planResp, nil
}

func (p *DetailData) ResolveData(data internal.GetPlanResponse) string {
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

	viewModel := DetailViewModel{
		ID:        data.ID,
		State:     string(data.State),
		MergeBase: data.MergeBase,
		Head:      data.Head,
		Add:       data.Add,
		Change:    data.Change,
		Destroy:   data.Destroy,
		Component: component,
		Changeset: changeset,
	}

	yamlData, err := yaml.Marshal(viewModel)
	if err != nil {
		return fmt.Sprintf("Error marshaling to YAML: %v", err)
	}

	return string(yamlData)
}

func (p *DetailData) KeyBindings(elem internal.GetPlanResponse) platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "l", Help: "View logs", Command: fmt.Sprintf("plans/%d/logs", elem.ID)},
		{Key: "c", Help: "View component", Command: fmt.Sprintf("components/%d", elem.ComponentID)},
	}
}
