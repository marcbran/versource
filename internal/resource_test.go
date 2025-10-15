package internal

import (
	"testing"

	"github.com/marcbran/versource/pkg/versource"
	"gorm.io/datatypes"
)

func TestApplyResourceMapping(t *testing.T) {
	resourceA := versource.StateResource{Resource: versource.Resource{Attributes: datatypes.JSON(`{"name":"a"}`)}}
	resourceB := versource.StateResource{Resource: versource.Resource{Attributes: datatypes.JSON(`{"name":"b"}`)}}
	resourceC := versource.StateResource{Resource: versource.Resource{Attributes: datatypes.JSON(`{"name":"c"}`)}}
	newResource := versource.StateResource{Resource: versource.Resource{Provider: "test", ResourceType: "test", Name: "new", Attributes: datatypes.JSON(`{"name":"new"}`)}}

	keepA := []datatypes.JSON{datatypes.JSON(`{"name":"a"}`)}
	dropA := []datatypes.JSON{datatypes.JSON(`{"name":"a"}`)}
	empty := []datatypes.JSON{}

	tests := []struct {
		name           string
		stateResources []versource.StateResource
		mapping        versource.ResourceMapping
		expected       []versource.StateResource
	}{
		{
			name:           "keep omitted - all resources kept",
			stateResources: []versource.StateResource{resourceA, resourceB, resourceC},
			mapping:        versource.ResourceMapping{Keep: nil, Drop: nil},
			expected:       []versource.StateResource{resourceA, resourceB, resourceC},
		},
		{
			name:           "keep empty - no resources kept",
			stateResources: []versource.StateResource{resourceA, resourceB, resourceC},
			mapping:        versource.ResourceMapping{Keep: &empty, Drop: nil},
			expected:       []versource.StateResource{},
		},
		{
			name:           "keep specific - only kept resources",
			stateResources: []versource.StateResource{resourceA, resourceB, resourceC},
			mapping:        versource.ResourceMapping{Keep: &keepA, Drop: nil},
			expected:       []versource.StateResource{resourceA},
		},
		{
			name:           "drop specific - dropped resources removed",
			stateResources: []versource.StateResource{resourceA, resourceB, resourceC},
			mapping:        versource.ResourceMapping{Keep: nil, Drop: &dropA},
			expected:       []versource.StateResource{resourceB, resourceC},
		},
		{
			name:           "keep and drop - keep takes precedence",
			stateResources: []versource.StateResource{resourceA, resourceB, resourceC},
			mapping:        versource.ResourceMapping{Keep: &keepA, Drop: &dropA},
			expected:       []versource.StateResource{resourceA},
		},
		{
			name:           "add resources - new resources added",
			stateResources: []versource.StateResource{resourceA, resourceB},
			mapping: versource.ResourceMapping{
				Keep: nil,
				Drop: nil,
				Add:  &[]versource.Resource{newResource.Resource},
			},
			expected: []versource.StateResource{
				resourceA,
				resourceB,
				newResource,
			},
		},
		{
			name:           "keep empty with add - only add resources",
			stateResources: []versource.StateResource{resourceA, resourceB, resourceC},
			mapping: versource.ResourceMapping{
				Keep: &empty,
				Drop: nil,
				Add:  &[]versource.Resource{newResource.Resource},
			},
			expected: []versource.StateResource{newResource},
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
