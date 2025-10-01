package apply

import (
	"context"
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/tui/platform"
)

type DetailData struct {
	facade  internal.Facade
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

func NewDetail(facade internal.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewViewDataViewport(NewDetailData(
			facade,
			params["applyID"],
		))
	}
}

func NewDetailData(facade internal.Facade, applyID string) *DetailData {
	return &DetailData{
		facade:  facade,
		applyID: applyID,
	}
}

func (p *DetailData) LoadData() (*internal.GetApplyResponse, error) {
	ctx := context.Background()

	applyIDUint, err := strconv.ParseUint(p.applyID, 10, 32)
	if err != nil {
		return nil, err
	}

	req := internal.GetApplyRequest{ApplyID: uint(applyIDUint)}

	applyResp, err := p.facade.GetApply(ctx, req)
	if err != nil {
		return nil, err
	}

	return applyResp, nil
}

func (p *DetailData) ResolveData(data internal.GetApplyResponse) DetailViewModel {
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
	if data.Plan.ID != 0 {
		var component *struct {
			ID   uint   `yaml:"id"`
			Name string `yaml:"name"`
		}
		if data.Plan.Component.ID != 0 {
			component = &struct {
				ID   uint   `yaml:"id"`
				Name string `yaml:"name"`
			}{
				ID:   data.Plan.Component.ID,
				Name: data.Plan.Component.Name,
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
			ID:          data.Plan.ID,
			State:       string(data.Plan.State),
			From:        data.Plan.From,
			To:          data.Plan.To,
			Add:         data.Plan.Add,
			Change:      data.Plan.Change,
			Destroy:     data.Plan.Destroy,
			ComponentID: data.Plan.ComponentID,
			Component:   component,
			ChangesetID: data.Plan.ChangesetID,
			Changeset:   changeset,
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
		ID:          data.ID,
		State:       string(data.State),
		PlanID:      data.PlanID,
		ChangesetID: data.ChangesetID,
		Plan:        plan,
		Changeset:   changeset,
	}
}

func (p *DetailData) KeyBindings(elem internal.GetApplyResponse) platform.KeyBindings {
	return platform.KeyBindings{
		{Key: "esc", Help: "View applies", Command: "applies"},
		{Key: "l", Help: "View logs", Command: fmt.Sprintf("applies/%s/logs", p.applyID)},
		{Key: "p", Help: "View plan", Command: fmt.Sprintf("plans/%d", elem.PlanID)},
		{Key: "c", Help: "View component", Command: fmt.Sprintf("components/%d", elem.Plan.ComponentID)},
	}
}
