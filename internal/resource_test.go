package internal

import (
	"testing"

	"gorm.io/datatypes"
)

func TestApplyResourceMapping(t *testing.T) {
	resourceA := StateResource{Resource: Resource{Attributes: datatypes.JSON(`{"name":"a"}`)}}
	resourceB := StateResource{Resource: Resource{Attributes: datatypes.JSON(`{"name":"b"}`)}}
	resourceC := StateResource{Resource: Resource{Attributes: datatypes.JSON(`{"name":"c"}`)}}

	keepA := []datatypes.JSON{datatypes.JSON(`{"name":"a"}`)}
	dropA := []datatypes.JSON{datatypes.JSON(`{"name":"a"}`)}
	empty := []datatypes.JSON{}

	tests := []struct {
		name           string
		stateResources []StateResource
		mapping        ResourceMapping
		expected       []StateResource
	}{
		{
			name:           "keep omitted - all resources kept",
			stateResources: []StateResource{resourceA, resourceB, resourceC},
			mapping:        ResourceMapping{Keep: nil, Drop: nil},
			expected:       []StateResource{resourceA, resourceB, resourceC},
		},
		{
			name:           "keep empty - no resources kept",
			stateResources: []StateResource{resourceA, resourceB, resourceC},
			mapping:        ResourceMapping{Keep: &empty, Drop: nil},
			expected:       []StateResource{},
		},
		{
			name:           "keep specific - only kept resources",
			stateResources: []StateResource{resourceA, resourceB, resourceC},
			mapping:        ResourceMapping{Keep: &keepA, Drop: nil},
			expected:       []StateResource{resourceA},
		},
		{
			name:           "drop specific - dropped resources removed",
			stateResources: []StateResource{resourceA, resourceB, resourceC},
			mapping:        ResourceMapping{Keep: nil, Drop: &dropA},
			expected:       []StateResource{resourceB, resourceC},
		},
		{
			name:           "keep and drop - keep takes precedence",
			stateResources: []StateResource{resourceA, resourceB, resourceC},
			mapping:        ResourceMapping{Keep: &keepA, Drop: &dropA},
			expected:       []StateResource{resourceA},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyResourceMapping(tt.stateResources, tt.mapping)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d resources, got %d", len(tt.expected), len(result))
			}
		})
	}
}
