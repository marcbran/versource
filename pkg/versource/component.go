package versource

import (
	"gorm.io/datatypes"
)

type Component struct {
	ID              uint            `gorm:"primarykey" json:"id" yaml:"id"`
	Name            string          `gorm:"not null;default:'';uniqueIndex" json:"name" yaml:"name"`
	ModuleVersion   ModuleVersion   `gorm:"foreignKey:ModuleVersionID" json:"moduleVersion" yaml:"moduleVersion"`
	ModuleVersionID uint            `json:"moduleVersionId" yaml:"moduleVersionId"`
	Variables       datatypes.JSON  `gorm:"type:jsonb" json:"variables" yaml:"variables"`
	Status          ComponentStatus `gorm:"default:Ready" json:"status" yaml:"status"`
}

type ComponentStatus string

const (
	ComponentStatusReady   ComponentStatus = "Ready"
	ComponentStatusDeleted ComponentStatus = "Deleted"
)

type ComponentChange struct {
	FromComponent *Component `json:"fromComponent,omitempty" yaml:"fromComponent,omitempty"`
	ToComponent   *Component `json:"toComponent,omitempty" yaml:"toComponent,omitempty"`
	ChangeType    ChangeType `json:"changeType" yaml:"changeType"`
	Plan          *Plan      `json:"plan,omitempty" yaml:"plan,omitempty"`
	FromCommit    string     `json:"fromCommit,omitempty" yaml:"fromCommit,omitempty"`
	ToCommit      string     `json:"toCommit,omitempty" yaml:"toCommit,omitempty"`
}

type ChangeType string

const (
	ChangeTypeCreated  ChangeType = "Created"
	ChangeTypeDeleted  ChangeType = "Deleted"
	ChangeTypeModified ChangeType = "Modified"
)

type GetComponentRequest struct {
	ComponentID   uint    `json:"componentId" yaml:"componentId"`
	ChangesetName *string `json:"changesetName,omitempty" yaml:"changesetName,omitempty"`
}

type GetComponentResponse struct {
	Component Component `json:"component" yaml:"component"`
}

type ListComponentsRequest struct {
	ModuleID        *uint   `json:"moduleId,omitempty" yaml:"moduleId,omitempty"`
	ModuleVersionID *uint   `json:"moduleVersionId,omitempty" yaml:"moduleVersionId,omitempty"`
	ChangesetName   *string `json:"changesetName,omitempty" yaml:"changesetName,omitempty"`
}

type ListComponentsResponse struct {
	Components []Component `json:"components" yaml:"components"`
}

type GetComponentChangeRequest struct {
	ComponentID   uint   `json:"componentId" yaml:"componentId"`
	ChangesetName string `json:"changesetName" yaml:"changesetName"`
}

type GetComponentChangeResponse struct {
	Change ComponentChange `json:"change" yaml:"change"`
}

type ListComponentChangesRequest struct {
	ChangesetName string `json:"changesetName" yaml:"changesetName"`
}

type ListComponentChangesResponse struct {
	Changes []ComponentChange `json:"changes" yaml:"changes"`
}

type CreateComponentRequest struct {
	ChangesetName string         `json:"changesetName" yaml:"changesetName"`
	ModuleID      uint           `json:"moduleId" yaml:"moduleId"`
	Name          string         `json:"name" yaml:"name"`
	Variables     map[string]any `json:"variables" yaml:"variables"`
}

type CreateComponentResponse struct {
	ID        uint           `json:"id" yaml:"id"`
	Name      string         `json:"name" yaml:"name"`
	Source    string         `json:"source" yaml:"source"`
	Version   string         `json:"version" yaml:"version"`
	Variables map[string]any `json:"variables" yaml:"variables"`
	PlanID    uint           `json:"planId" yaml:"planId"`
}

type UpdateComponentRequest struct {
	ComponentID   uint            `json:"componentId" yaml:"componentId"`
	ChangesetName string          `json:"changesetName" yaml:"changesetName"`
	ModuleID      *uint           `json:"moduleId,omitempty" yaml:"moduleId,omitempty"`
	Variables     *map[string]any `json:"variables,omitempty" yaml:"variables,omitempty"`
}

type UpdateComponentResponse struct {
	ID        uint           `json:"id" yaml:"id"`
	Name      string         `json:"name" yaml:"name"`
	Source    string         `json:"source" yaml:"source"`
	Version   string         `json:"version" yaml:"version"`
	Variables map[string]any `json:"variables" yaml:"variables"`
	PlanID    uint           `json:"planId" yaml:"planId"`
}

type DeleteComponentRequest struct {
	ComponentID   uint   `json:"componentId" yaml:"componentId"`
	ChangesetName string `json:"changesetName" yaml:"changesetName"`
}

type DeleteComponentResponse struct {
	ID        uint            `json:"id" yaml:"id"`
	Name      string          `json:"name" yaml:"name"`
	Source    string          `json:"source" yaml:"source"`
	Version   string          `json:"version" yaml:"version"`
	Variables map[string]any  `json:"variables" yaml:"variables"`
	Status    ComponentStatus `json:"status" yaml:"status"`
	PlanID    uint            `json:"planId" yaml:"planId"`
}

type RestoreComponentRequest struct {
	ComponentID   uint   `json:"componentId" yaml:"componentId"`
	ChangesetName string `json:"changesetName" yaml:"changesetName"`
}

type RestoreComponentResponse struct {
	ID        uint            `json:"id" yaml:"id"`
	Name      string          `json:"name" yaml:"name"`
	Source    string          `json:"source" yaml:"source"`
	Version   string          `json:"version" yaml:"version"`
	Variables map[string]any  `json:"variables" yaml:"variables"`
	Status    ComponentStatus `json:"status" yaml:"status"`
	PlanID    uint            `json:"planId" yaml:"planId"`
}
