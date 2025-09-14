package component

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/tui/platform"
)

type EditComponentData struct {
	facade        internal.Facade
	componentID   string
	changesetName string
}

func NewEdit(facade internal.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewEditor(&EditComponentData{
			facade:        facade,
			componentID:   params["componentID"],
			changesetName: params["changesetName"],
		})
	}
}

func (e *EditComponentData) GetInitialValue() internal.UpdateComponentRequest {
	componentID := uint(0)
	if e.componentID != "" {
		if id, err := strconv.ParseUint(e.componentID, 10, 32); err == nil {
			componentID = uint(id)
		}
	}

	ctx := context.Background()
	componentResp, err := e.facade.GetComponent(ctx, internal.GetComponentRequest{ComponentID: componentID, Changeset: &e.changesetName})
	if err != nil {
		return internal.UpdateComponentRequest{
			ComponentID: componentID,
			Changeset:   e.changesetName,
			ModuleID:    nil,
			Variables:   nil,
		}
	}

	var variables map[string]any
	if componentResp.Component.Variables != nil {
		json.Unmarshal(componentResp.Component.Variables, &variables)
	}

	return internal.UpdateComponentRequest{
		ComponentID: componentID,
		Changeset:   e.changesetName,
		ModuleID:    &componentResp.Component.ModuleVersion.Module.ID,
		Variables:   &variables,
	}
}

func (e *EditComponentData) SaveData(ctx context.Context, data internal.UpdateComponentRequest) (string, error) {
	if data.ComponentID == 0 {
		return "", fmt.Errorf("component ID is required")
	}

	if data.Changeset == "" {
		return "", fmt.Errorf("changeset is required")
	}

	_, err := e.facade.UpdateComponent(ctx, data)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("changesets/%s/components", data.Changeset), nil
}
