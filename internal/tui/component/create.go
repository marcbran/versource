package component

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/marcbran/versource/internal"
	"github.com/marcbran/versource/internal/tui/platform"
)

type CreateComponentData struct {
	facade        internal.Facade
	moduleID      string
	changesetName string
}

func NewCreateComponent(facade internal.Facade) func(params map[string]string) platform.Page {
	return func(params map[string]string) platform.Page {
		return platform.NewEditor(&CreateComponentData{
			facade:        facade,
			moduleID:      params["module-id"],
			changesetName: params["changesetName"],
		})
	}
}

func (c *CreateComponentData) GetInitialValue() (internal.CreateComponentRequest, error) {
	moduleID := uint(0)
	if c.moduleID != "" {
		id, err := strconv.ParseUint(c.moduleID, 10, 32)
		if err != nil {
			return internal.CreateComponentRequest{}, err
		}
		moduleID = uint(id)
	}

	changesetName := c.changesetName
	if changesetName == "" {
		changesetName = generateDefaultChangesetName("create")
	}

	return internal.CreateComponentRequest{
		ChangesetName: changesetName,
		ModuleID:      moduleID,
		Name:          "",
		Variables:     make(map[string]any),
	}, nil
}

func (c *CreateComponentData) SaveData(ctx context.Context, data internal.CreateComponentRequest) (string, error) {
	if data.ChangesetName == "" {
		return "", fmt.Errorf("changeset is required")
	}

	if data.ModuleID == 0 {
		return "", fmt.Errorf("moduleId is required")
	}

	if data.Name == "" {
		return "", fmt.Errorf("name is required")
	}

	_, err := c.facade.CreateComponent(ctx, data)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("changesets/%s/changes", data.ChangesetName), nil
}

func generateDefaultChangesetName(prefix string) string {
	now := time.Now()
	dateStr := now.Format("060102")

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	chars := "abcdefghijklmnopqrstuvwxyz0123456789"
	randomStr := make([]byte, 6)
	for i := range randomStr {
		randomStr[i] = chars[r.Intn(len(chars))]
	}

	return fmt.Sprintf("%s-%s-%s", prefix, dateStr, string(randomStr))
}
