package apply

import (
	"context"
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal/tui/platform"
	"github.com/marcbran/versource/pkg/versource"
)

type DetailData struct {
	facade  versource.Facade
	applyID string
}

type DetailViewModel struct {
	ID          uint   `yaml:"id"`
	State       string `yaml:"state"`
	PlanID      uint   `yaml:"planId"`
	ChangesetID uint   `yaml:"changesetId"`
	Plan        *struct {
		ID          uint   `yaml:"id"`
		State       string `yaml:"state"`
		From        string `yaml:"from"`
		To          string `yaml:"to"`
		Add         *int   `yaml:"add,omitempty"`
		Change      *int   `yaml:"change,omitempty"`
		Destroy     *int   `yaml:"destroy,omitempty"`
		ComponentID uint   `yaml:"componentId"`
		Component   *struct {
			ID   uint   `yaml:"id"`
			Name string `yaml:"name"`
		} `yaml:"component,omitempty"`
		ChangesetID uint `yaml:"changesetId"`
		Changeset   *struct {
			ID   uint   `yaml:"id"`
			Name string `yaml:"name"`
		} `yaml:"changeset,omitempty"`
	} `yaml:"plan,omitempty"`
	Changeset *struct {
		ID   uint   `yaml:"id"`
		Name string `yaml:"name"`
	} `yaml:"changeset,omitempty"`
}

func NewDetail(facade versource.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewViewDataViewport(NewDetailData(
			facade,
			params["applyID"],
		))
	}
}

func NewDetailData(facade versource.Facade, applyID string) *DetailData {
	return &DetailData{
		facade:  facade,
		applyID: applyID,
	}
}

func (p *DetailData) LoadData() (*versource.GetApplyResponse, error) {
	ctx := context.Background()

	applyIDUint, err := strconv.ParseUint(p.applyID, 10, 32)
	if err != nil {
		return nil, err
	}

	req := versource.GetApplyRequest{ApplyID: uint(applyIDUint)}

	applyResp, err := p.facade.GetApply(ctx, req)
	if err != nil {
		return nil, err
	}

	return applyResp, nil
}

func (p *DetailData) ResolveData(data versource.GetApplyResponse) DetailViewModel {
	var plan *struct {
		ID          uint   `yaml:"id"`
		State       string `yaml:"state"`
		From        string `yaml:"from"`
		To          string `yaml:"to"`
		Add         *int   `yaml:"add,omitempty"`
		Change      *int   `yaml:"change,omitempty"`
		Destroy     *int   `yaml:"destroy,omitempty"`
		ComponentID uint   `yaml:"componentId"`
		Component   *struct {
			ID   uint   `yaml:"id"`
			Name string `yaml:"name"`
		} `yaml:"component,omitempty"`
		ChangesetID uint `yaml:"changesetId"`
		Changeset   *struct {
			ID   uint   `yaml:"id"`
			Name string `yaml:"name"`
		} `yaml:"changeset,omitempty"`
	}
	if data.Apply.Plan.ID != 0 {
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
		if data.Apply.Plan.Changeset.ID != 0 {
			changeset = &struct {
				ID   uint   `yaml:"id"`
				Name string `yaml:"name"`
			}{
				ID:   data.Apply.Plan.Changeset.ID,
				Name: data.Apply.Plan.Changeset.Name,
			}
		}

		plan = &struct {
			ID          uint   `yaml:"id"`
			State       string `yaml:"state"`
			From        string `yaml:"from"`
			To          string `yaml:"to"`
			Add         *int   `yaml:"add,omitempty"`
			Change      *int   `yaml:"change,omitempty"`
			Destroy     *int   `yaml:"destroy,omitempty"`
			ComponentID uint   `yaml:"componentId"`
			Component   *struct {
				ID   uint   `yaml:"id"`
				Name string `yaml:"name"`
			} `yaml:"component,omitempty"`
			ChangesetID uint `yaml:"changesetId"`
			Changeset   *struct {
				ID   uint   `yaml:"id"`
				Name string `yaml:"name"`
			} `yaml:"changeset,omitempty"`
		}{
			ID:          data.Apply.Plan.ID,
			State:       string(data.Apply.Plan.State),
			From:        data.Apply.Plan.From,
			To:          data.Apply.Plan.To,
			Add:         data.Apply.Plan.Add,
			Change:      data.Apply.Plan.Change,
			Destroy:     data.Apply.Plan.Destroy,
			ComponentID: data.Apply.Plan.ComponentID,
			Component:   component,
			ChangesetID: data.Apply.Plan.ChangesetID,
			Changeset:   changeset,
		}
	}

	var changeset *struct {
		ID   uint   `yaml:"id"`
		Name string `yaml:"name"`
	}
	if data.Apply.Changeset.ID != 0 {
		changeset = &struct {
			ID   uint   `yaml:"id"`
			Name string `yaml:"name"`
		}{
			ID:   data.Apply.Changeset.ID,
			Name: data.Apply.Changeset.Name,
		}
	}

	return DetailViewModel{
		ID:          data.Apply.ID,
		State:       string(data.Apply.State),
		PlanID:      data.Apply.PlanID,
		ChangesetID: data.Apply.ChangesetID,
		Plan:        plan,
		Changeset:   changeset,
	}
}

func (p *DetailData) KeyBindings(elem versource.GetApplyResponse) platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "esc", Help: "View applies", Command: "applies"},
		{Key: "l", Help: "View logs", Command: fmt.Sprintf("applies/%s/logs", p.applyID)},
		{Key: "p", Help: "View plan", Command: fmt.Sprintf("plans/%d", elem.Apply.PlanID)},
		{Key: "c", Help: "View component", Command: fmt.Sprintf("components/%d", elem.Apply.Plan.ComponentID)},
	}
}
