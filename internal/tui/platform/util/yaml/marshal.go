package yaml

import (
	"reflect"
	"strconv"

	"gopkg.in/yaml.v3"
)

func Marshal(v interface{}) ([]byte, error) {
	node, err := marshalToNode(v)
	if err != nil {
		return nil, err
	}
	return yaml.Marshal(node)
}

func marshalToNode(v interface{}) (*yaml.Node, error) {
	rv := reflect.ValueOf(v)

	if rv.Kind() == reflect.Ptr && rv.IsNil() {
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: "null",
		}, nil
	}

	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Struct:
		return marshalStructToNode(rv)
	case reflect.Map:
		return marshalMapToNode(rv)
	case reflect.Slice:
		return marshalSliceToNode(rv)
	default:
		return marshalScalarToNode(rv)
	}
}

func marshalStructToNode(rv reflect.Value) (*yaml.Node, error) {
	node := &yaml.Node{
		Kind:    yaml.MappingNode,
		Content: []*yaml.Node{},
	}

	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		if !field.IsExported() {
			continue
		}

		yamlTag := field.Tag.Get("yaml")
		if yamlTag == "-" {
			continue
		}

		fieldName := field.Name
		if yamlTag != "" {
			parts := splitYamlTag(yamlTag)
			if parts[0] != "" {
				fieldName = parts[0]
			}
		}

		fieldValue := rv.Field(i)
		fieldNode, err := marshalToNode(fieldValue.Interface())
		if err != nil {
			return nil, err
		}

		keyNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: fieldName,
		}

		node.Content = append(node.Content, keyNode, fieldNode)
	}

	return node, nil
}

func marshalMapToNode(rv reflect.Value) (*yaml.Node, error) {
	node := &yaml.Node{
		Kind:    yaml.MappingNode,
		Content: []*yaml.Node{},
	}

	for _, key := range rv.MapKeys() {
		keyNode, err := marshalToNode(key.Interface())
		if err != nil {
			return nil, err
		}

		valueNode, err := marshalToNode(rv.MapIndex(key).Interface())
		if err != nil {
			return nil, err
		}

		node.Content = append(node.Content, keyNode, valueNode)
	}

	return node, nil
}

func marshalSliceToNode(rv reflect.Value) (*yaml.Node, error) {
	node := &yaml.Node{
		Kind:    yaml.SequenceNode,
		Content: []*yaml.Node{},
	}

	for i := 0; i < rv.Len(); i++ {
		itemNode, err := marshalToNode(rv.Index(i).Interface())
		if err != nil {
			return nil, err
		}
		node.Content = append(node.Content, itemNode)
	}

	return node, nil
}

func marshalScalarToNode(rv reflect.Value) (*yaml.Node, error) {
	node := &yaml.Node{
		Kind: yaml.ScalarNode,
	}

	switch rv.Kind() {
	case reflect.String:
		node.Value = rv.String()
		if node.Value == "" {
			node.Style = 0
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		node.Value = formatInt(rv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		node.Value = formatUint(rv.Uint())
	case reflect.Float32, reflect.Float64:
		node.Value = formatFloat(rv.Float())
	case reflect.Bool:
		node.Value = formatBool(rv.Bool())
	default:
		node.Value = rv.String()
	}

	return node, nil
}

func splitYamlTag(tag string) []string {
	parts := make([]string, 0, 3)
	start := 0
	for i, c := range tag {
		if c == ',' {
			parts = append(parts, tag[start:i])
			start = i + 1
		}
	}
	parts = append(parts, tag[start:])
	return parts
}

func formatInt(i int64) string {
	return strconv.FormatInt(i, 10)
}

func formatUint(u uint64) string {
	return strconv.FormatUint(u, 10)
}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'g', -1, 64)
}

func formatBool(b bool) string {
	return strconv.FormatBool(b)
}
