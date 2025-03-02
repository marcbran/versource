package internal

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/dolthub/driver"
	"path"
	"strings"
)

type Item struct {
	Uid       string            `json:"uid"`
	Title     string            `json:"title"`
	Arg       string            `json:"arg"`
	Variables map[string]string `json:"variables"`
}

type ItemList struct {
	Items []Item `json:"items"`
}

func List(ctx context.Context, dataDir string, view string) (ItemList, error) {
	dbDir := path.Join(dataDir, "db")
	db, err := sql.Open("dolt", fmt.Sprintf("file://%s?database=%s&commitname=%s&commitemail=%s", dbDir, "versource", "none", "none"))
	if err != nil {
		return ItemList{}, err
	}

	if view == "" {
		list, err := listViews(ctx, db)
		if err != nil {
			return ItemList{}, err
		}
		return list, nil
	}

	list, err := listViewItems(ctx, db, view)
	if err != nil {
		return ItemList{}, err
	}
	return list, nil
}

func listViews(ctx context.Context, db *sql.DB) (ItemList, error) {
	rows, err := db.QueryContext(ctx, "SHOW FULL TABLES IN `versource` WHERE Table_type = 'VIEW' AND Tables_in_versource LIKE '%_items' AND Tables_in_versource <> 'view_items'")
	if err != nil {
		return ItemList{}, err
	}
	defer rows.Close()

	items := []Item{}
	for rows.Next() {
		var name, typ string
		err := rows.Scan(&name, &typ)
		if err != nil {
			return ItemList{}, err
		}
		title := strings.TrimSuffix(strings.ReplaceAll(name, "_", "-"), "-items")
		items = append(items, Item{
			Uid:   fmt.Sprintf("view-%s", title),
			Title: title,
			Arg:   "",
			Variables: map[string]string{
				"view": title,
			},
		})
	}
	list := ItemList{
		Items: items,
	}
	return list, nil
}

func listViewItems(ctx context.Context, db *sql.DB, view string) (ItemList, error) {
	name := strings.ReplaceAll(fmt.Sprintf("%s-items", view), "-", "_")
	rows, err := db.QueryContext(ctx, fmt.Sprintf("SELECT uid, title, arg FROM `%s`", name))
	if err != nil {
		return ItemList{}, err
	}
	defer rows.Close()

	items := []Item{}
	for rows.Next() {
		var uid, title, arg string
		err := rows.Scan(&uid, &title, &arg)
		if err != nil {
			return ItemList{}, err
		}
		items = append(items, Item{
			Uid:   uid,
			Title: title,
			Arg:   arg,
		})
	}
	list := ItemList{
		Items: items,
	}
	return list, nil
}
