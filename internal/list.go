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
	Subtitle     string            `json:"subtitle"`
	Match        string            `json:"match"`
	Icon         Icon              `json:"icon"`
	Arg          []string          `json:"arg"`
	Autocomplete string            `json:"autocomplete,omitempty"`
	Variables    map[string]string `json:"variables"`
}

type Icon struct {
	Path string `json:"path"`
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

func (r Resource) Id() string {
	return joinNonEmpty([]string{r.Provider, r.ResourceType, r.ProviderAlias, r.Namespace, r.Name}, "/")
}

func (r Resource) FullResourceType() string {
	return joinNonEmpty([]string{r.Provider, r.ResourceType}, " ")
}

func (r Resource) FullName() string {
	return joinNonEmpty([]string{r.ProviderAlias, r.Namespace, r.Name}, "/")
}

type ResourceIds struct {
	Uuid             string `json:"uuid"`
	Id               string `json:"id"`
	FullResourceType string `json:"full_resource_type"`
	FullName         string `json:"full_name"`
}

func List(ctx context.Context, configDir, dataDir, query, resource string) (FilterList, error) {
	dbDir := path.Join(dataDir, "db")
	db, err := sql.Open("dolt", fmt.Sprintf("file://%s?database=%s&commitname=%s&commitemail=%s", dbDir, "versource", "none", "none"))
	if err != nil {
		return FilterList{}, err
	}
	list, err := listResources(ctx, db, resource, configDir)
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

func listResources(ctx context.Context, db *sql.DB, resource string, configDir string) (FilterList, error) {
	if resource != "" {
		return listResourceProjections(ctx, db, configDir, resource)
	}
	resources, err := listAllResources(ctx, db, configDir)
	if err != nil {
		return FilterList{}, err
	}
	return resources, nil
}

func listResourceProjections(ctx context.Context, db *sql.DB, configDir string, resourceId string) (FilterList, error) {
	r, err := getResource(ctx, db, resourceId)
	if err != nil {
		return FilterList{}, err
	}
	resources, err := listResourceProjectionsForResource(configDir, r)
	if err != nil {
		return FilterList{}, err
	}
	items, err := itemizeResources(configDir, resources)
	if err != nil {
		return FilterList{}, err
	}
	list := FilterList{
		Items: items,
	}
	return list, nil
}

func listAllResources(ctx context.Context, db *sql.DB, configDir string) (FilterList, error) {
	resourceIds, err := listAllResourceIds(ctx, db)
	if err != nil {
		return FilterList{}, err
	}
	list := FilterList{
		Items: itemizeResourceIds(configDir, resourceIds),
	}
	return list, nil
}

func listAllResourceIds(ctx context.Context, db *sql.DB) ([]ResourceIds, error) {
	rows, err := db.QueryContext(
		ctx,
		`SELECT
		  uuid,
		  CONCAT_WS(
			'/',
			NULLIF(provider, ''),
			NULLIF(resource_type, ''),
			NULLIF(NULLIF(provider_alias, ''), 'default'),
			NULLIF(namespace, ''),
			NULLIF(name, '')
		  ) AS id,
		  CONCAT_WS(
			' ',
			NULLIF(provider, ''),
			NULLIF(resource_type, '')
		  ) AS full_resource_type,
		  CONCAT_WS(
			'/',
			NULLIF(NULLIF(provider_alias, ''), 'default'),
			NULLIF(namespace, ''),
			NULLIF(name, '')
		  ) AS full_name
		FROM resources`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var resourceIds []ResourceIds
	for rows.Next() {
		var uuid, id, fullResourceType, fullName string
		err := rows.Scan(&uuid, &id, &fullResourceType, &fullName)
		if err != nil {
			return nil, err
		}
		resourceIds = append(resourceIds, ResourceIds{
			Uuid:             uuid,
			Id:               id,
			FullResourceType: fullResourceType,
			FullName:         fullName,
		})
	}
	return resourceIds, nil
}

func getResource(ctx context.Context, db *sql.DB, resourceId string) (Resource, error) {
	rows, err := db.QueryContext(
		ctx,
		`SELECT *
		FROM resources
		WHERE
		CONCAT_WS(
			'/',
			NULLIF(provider, ''),
			NULLIF(resource_type, ''),
			NULLIF(NULLIF(provider_alias, ''), 'default'),
			NULLIF(namespace, ''),
			NULLIF(name, '')
		) = ?
		LIMIT 1`,
		resourceId,
	)
	if err != nil {
		return Resource{}, err
	}
	defer rows.Close()

	if !rows.Next() {
		return Resource{}, fmt.Errorf("cannot find resource with id %s", resourceId)
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

func itemizeResources(configDir string, resources []Resource) ([]Item, error) {
	var items []Item
	for _, resource := range resources {
		b, err := json.Marshal(resource)
		if err != nil {
			return nil, err
		}
		items = append(items, Item{
			Uid:      resource.Uuid,
			Title:    resource.FullName(),
			Subtitle: resource.FullResourceType(),
			Match:    resource.Id(),
			Icon: Icon{
				Path: path.Join(configDir, "icons", fmt.Sprintf("%s.svg", resource.Provider)),
			},
			Variables: map[string]string{
				"resource": string(b),
			},
		})
	}
	return items, nil
}

func itemizeResourceIds(configDir string, resourceIds []ResourceIds) []Item {
	var items []Item
	for _, resourceId := range resourceIds {
		items = append(items, Item{
			Uid:      resourceId.Uuid,
			Title:    resourceId.FullName,
			Subtitle: resourceId.FullResourceType,
			Match:    resourceId.Id,
			Icon: Icon{
				Path: path.Join(configDir, "icons", fmt.Sprintf("%s.svg", strings.Split(resourceId.FullResourceType, " ")[0])),
			},
			Variables: map[string]string{
				"resource": resourceId.Id,
			},
		})
	}
	return items
}

func filter(filter FilterList, query string) (FilterList, error) {
	if len(filter.Items) < 2 {
		return filter, nil
	}

	var matches []string
	matchesToItem := make(map[string]Item)
	for _, item := range filter.Items {
		matches = append(matches, item.Match)
		matchesToItem[item.Match] = item
	}

	cmd := exec.Command("fzf", "-f", query)
	cmd.Stdin = strings.NewReader(strings.Join(matches, "\n"))
	cmd.Stderr = os.Stderr
	b, err := cmd.Output()
	if err != nil {
		return FilterList{}, err
	}
	filteredMatches := strings.Split(string(b), "\n")

	var filteredItems []Item
	for _, match := range filteredMatches {
		filteredItems = append(filteredItems, matchesToItem[match])
	}
	return FilterList{
		Items:     filteredItems,
		Variables: filter.Variables,
		Rerun:     filter.Rerun,
	}, nil
}

func joinNonEmpty(elems []string, sep string) string {
	var nonEmptyElems []string
	for _, e := range elems {
		if e == "" {
			continue
		}
		nonEmptyElems = append(nonEmptyElems, e)
	}
	return strings.Join(nonEmptyElems, sep)
}
