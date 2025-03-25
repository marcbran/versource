package internal

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/dolthub/driver"
	"github.com/google/go-jsonnet"
	"os"
	"os/exec"
	"path"
	"strings"
)

type FilterList struct {
	Items     []Item            `json:"items"`
	Variables map[string]string `json:"variables"`
	Rerun     float32           `json:"rerun"`
}

type Item struct {
	Uid          string            `json:"uid"`
	Title        string            `json:"title"`
	Arg          []string          `json:"arg"`
	Autocomplete string            `json:"autocomplete,omitempty"`
	Variables    map[string]string `json:"variables"`
}

type Resource struct {
	Uuid          string         `json:"uuid"`
	Provider      string         `json:"provider"`
	ProviderAlias string         `json:"providerAlias"`
	ResourceType  string         `json:"resourceType"`
	Namespace     string         `json:"namespace"`
	Name          string         `json:"name"`
	Data          map[string]any `json:"data"`
}

type ResourceTitle struct {
	Uuid  string `json:"uuid"`
	Title string `json:"title"`
}

func List(ctx context.Context, configDir, dataDir, query, resource string) (FilterList, error) {
	dbDir := path.Join(dataDir, "db")
	db, err := sql.Open("dolt", fmt.Sprintf("file://%s?database=%s&commitname=%s&commitemail=%s", dbDir, "versource", "none", "none"))
	if err != nil {
		return FilterList{}, err
	}
	list, err := listResources(ctx, db, query, resource, configDir)
	if err != nil {
		return FilterList{}, err
	}
	list, err = filter(list, query)
	if err != nil {
		return FilterList{}, err
	}
	if list.Variables == nil {
		list.Variables = make(map[string]string)
	}
	list.Variables["query"] = query
	return list, nil
}

func listResources(ctx context.Context, db *sql.DB, query string, resource string, configDir string) (FilterList, error) {
	if strings.HasSuffix(query, " ") {
		if strings.TrimSuffix(query, " ") == resource {
			return listResourceProjections(ctx, db, configDir, resource)
		}
		return FilterList{
			Variables: map[string]string{
				"resource": strings.TrimSuffix(query, " "),
			},
			Rerun: 0.01,
		}, nil
	}
	resources, err := listAllResources(ctx, db)
	if err != nil {
		return FilterList{}, err
	}
	return resources, nil
}

func listResourceProjections(ctx context.Context, db *sql.DB, configDir string, resourceTitle string) (FilterList, error) {
	r, err := getResource(ctx, db, resourceTitle)
	if err != nil {
		return FilterList{}, err
	}
	resources, err := listResourceProjectionsForResource(configDir, r)
	if err != nil {
		return FilterList{}, err
	}
	items, err := itemizeResources(resources)
	if err != nil {
		return FilterList{}, err
	}
	list := FilterList{
		Items: items,
	}
	return list, nil
}

func listAllResources(ctx context.Context, db *sql.DB) (FilterList, error) {
	resourceTitles, err := listAllResourceTitles(ctx, db)
	if err != nil {
		return FilterList{}, err
	}
	list := FilterList{
		Items: itemizeResourceTitles(resourceTitles),
	}
	return list, nil
}

func listAllResourceTitles(ctx context.Context, db *sql.DB) ([]ResourceTitle, error) {
	rows, err := db.QueryContext(
		ctx,
		`SELECT
		  uuid,
		  CONCAT_WS(
			'/',
			NULLIF(provider, ''),
			NULLIF(provider_alias, ''),
			NULLIF(resource_type, ''),
			NULLIF(namespace, ''),
			NULLIF(name, '')
		  ) AS title
		FROM resources`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resourceTitles []ResourceTitle
	for rows.Next() {
		var uuid, title string
		err := rows.Scan(&uuid, &title)
		if err != nil {
			return nil, err
		}
		resourceTitles = append(resourceTitles, ResourceTitle{
			Uuid:  uuid,
			Title: title,
		})
	}
	return resourceTitles, nil
}

func getResource(ctx context.Context, db *sql.DB, resourceTitle string) (Resource, error) {
	rows, err := db.QueryContext(
		ctx,
		`SELECT *
		FROM resources
		WHERE
		CONCAT_WS(
			'/',
			NULLIF(provider, ''),
			NULLIF(provider_alias, ''),
			NULLIF(resource_type, ''),
			NULLIF(namespace, ''),
			NULLIF(name, '')
		) = ?
		LIMIT 1`,
		resourceTitle,
	)
	if err != nil {
		return Resource{}, err
	}
	defer rows.Close()

	if !rows.Next() {
		return Resource{}, fmt.Errorf("cannot find resource with title %s", resourceTitle)
	}
	var uuid, provider, providerAlias, resourceType, namespace, name, dataString string
	err = rows.Scan(&uuid, &provider, &providerAlias, &resourceType, &namespace, &name, &dataString)
	if err != nil {
		return Resource{}, err
	}
	var data map[string]any
	err = json.Unmarshal([]byte(dataString), &data)
	if err != nil {
		return Resource{}, err
	}
	r := Resource{
		Uuid:          uuid,
		Provider:      provider,
		ProviderAlias: providerAlias,
		ResourceType:  resourceType,
		Namespace:     namespace,
		Name:          name,
		Data:          data,
	}
	return r, nil
}

func listResourceProjectionsForResource(configDir string, resource Resource) ([]Resource, error) {
	vendorDir := path.Join(configDir, "vendor")
	vm := jsonnet.MakeVM()
	vm.Importer(&jsonnet.FileImporter{
		JPaths: []string{vendorDir, configDir},
	})
	b, err := json.Marshal(resource)
	if err != nil {
		return nil, err
	}
	res, err := vm.EvaluateAnonymousSnippet("main.jsonnet", fmt.Sprintf(`
        local main = import 'versource/main.libsonnet';
        local plugins = import 'plugins/main.libsonnet';
        local resource = %s;
        main.project(resource, plugins)
	`, string(b)))
	if err != nil {
		return nil, err
	}
	var projections []Resource
	err = json.Unmarshal([]byte(res), &projections)
	if err != nil {
		return nil, err
	}
	return projections, nil
}

func itemizeResources(resources []Resource) ([]Item, error) {
	var items []Item
	for _, resource := range resources {
		title := fmt.Sprintf("%s/%s/%s/%s", resource.Provider, resource.ProviderAlias, resource.Namespace, resource.Name)
		b, err := json.Marshal(resource)
		if err != nil {
			return nil, err
		}
		items = append(items, Item{
			Uid:          resource.Uuid,
			Title:        title,
			Autocomplete: fmt.Sprintf("%s ", title),
			Variables: map[string]string{
				"resource": string(b),
			},
		})
	}
	return items, nil
}

func itemizeResourceTitles(resourceTitles []ResourceTitle) []Item {
	var items []Item
	for _, resourceTitle := range resourceTitles {
		items = append(items, Item{
			Uid:          resourceTitle.Uuid,
			Title:        resourceTitle.Title,
			Autocomplete: fmt.Sprintf("%s ", resourceTitle.Title),
			Variables: map[string]string{
				"resource": resourceTitle.Title,
			},
		})
	}
	return items
}

func filter(filter FilterList, query string) (FilterList, error) {
	// TODO handle the case where a resource is selected in the query but another list is shown to be filtered
	if strings.HasSuffix(query, " ") {
		return filter, nil
	}
	if len(filter.Items) < 2 {
		return filter, nil
	}

	var titles []string
	titlesToItem := make(map[string]Item)
	for _, item := range filter.Items {
		titles = append(titles, item.Title)
		titlesToItem[item.Title] = item
	}

	cmd := exec.Command("fzf", "-f", query)
	cmd.Stdin = strings.NewReader(strings.Join(titles, "\n"))
	cmd.Stderr = os.Stderr
	b, err := cmd.Output()
	if err != nil {
		return FilterList{}, err
	}
	filteredTitles := strings.Split(string(b), "\n")

	var filteredItems []Item
	for _, title := range filteredTitles {
		filteredItems = append(filteredItems, titlesToItem[title])
	}
	return FilterList{
		Items:     filteredItems,
		Variables: filter.Variables,
		Rerun:     filter.Rerun,
	}, nil
}
