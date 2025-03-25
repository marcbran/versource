package internal

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/dolthub/driver"
	"os/exec"
	"path"
	"strings"
)

func Run(ctx context.Context, configDir, dataDir, resourceOrResourceTitle string) error {
	dbDir := path.Join(dataDir, "db")
	db, err := sql.Open("dolt", fmt.Sprintf("file://%s?database=%s&commitname=%s&commitemail=%s", dbDir, "versource", "none", "none"))
	if err != nil {
		return err
	}

	if strings.HasPrefix(resourceOrResourceTitle, "{") {
		var resource Resource
		err := json.Unmarshal([]byte(resourceOrResourceTitle), &resource)
		if err != nil {
			return err
		}
		return runResource(resource)
	}

	return runResourceTitle(ctx, db, configDir, resourceOrResourceTitle)
}

func runResource(resource Resource) error {
	switch resource.ResourceType {
	case "Page":
		p, err := exec.LookPath("open")
		if err != nil {
			return err
		}
		url, ok := resource.Data["url"].(string)
		if !ok {
			return errors.New("cannot find url")
		}
		cmd := exec.Command(p, url)
		err = cmd.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

func runResourceTitle(ctx context.Context, db *sql.DB, configDir, resourceTitle string) error {
	resource, err := getResource(ctx, db, resourceTitle)
	if err != nil {
		return err
	}
	projections, err := listResourceProjectionsForResource(configDir, resource)
	if err != nil {
		return err
	}
	if len(projections) == 0 {
		return nil
	}
	firstResource := projections[0]
	err = runResource(firstResource)
	if err != nil {
		return err
	}
	return nil
}
