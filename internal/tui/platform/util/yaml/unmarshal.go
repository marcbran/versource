package yaml

import (
	"gopkg.in/yaml.v3"
)

func Unmarshal(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}
