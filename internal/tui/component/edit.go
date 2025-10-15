package component

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/marcbran/versource/internal/tui/platform"
	"github.com/marcbran/versource/pkg/versource"
)

type EditComponentData struct {
	facade        versource.Facade
	componentID   string
	changesetName string
}

func NewEdit(facade versource.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewEditor(&EditComponentData{
			facade:        facade,
			componentID:   params["componentID"],
			changesetName: params["changesetName"],
		})
	}
}

func (e *EditComponentData) GetInitialValue() (versource.UpdateComponentRequest, error) {
	componentID := uint(0)
	if e.componentID != "" {
		id, err := strconv.ParseUint(e.componentID, 10, 32)
		if err != nil {
			return versource.UpdateComponentRequest{}, err
		}
		componentID = uint(id)
	}

	ctx := context.Background()
	componentResp, err := e.facade.GetComponent(ctx, versource.GetComponentRequest{ComponentID: componentID, ChangesetName: &e.changesetName})
	if err != nil {
		return versource.UpdateComponentRequest{}, err
	}

	var variables map[string]any
	if componentResp.Component.Variables != nil {
		err := json.Unmarshal(componentResp.Component.Variables, &variables)
		if err != nil {
			return versource.UpdateComponentRequest{}, err
		}
	}

	changesetName := e.changesetName
	if changesetName == "" {
		changesetName = generateDefaultChangesetName(fmt.Sprintf("%s-update", componentResp.Component.Name))
	}

	return versource.UpdateComponentRequest{
		ComponentID:   componentID,
		ChangesetName: changesetName,
		ModuleID:      &componentResp.Component.ModuleVersion.Module.ID,
		Variables:     &variables,
	}, nil
}

func (e *EditComponentData) SaveData(ctx context.Context, data versource.UpdateComponentRequest) (string, error) {
	if data.ComponentID == 0 {
		return "", fmt.Errorf("component ID is required")
	}

	if data.ChangesetName == "" {
		return "", fmt.Errorf("changeset is required")
	}

	_, err := e.facade.UpdateComponent(ctx, data)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("changesets/%s/changes", data.ChangesetName), nil
}
