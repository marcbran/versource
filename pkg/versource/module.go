package versource

import (
	"gorm.io/datatypes"
)

type Module struct {
	ID           uint   `gorm:"primarykey" json:"id" yaml:"id"`
	Name         string `gorm:"uniqueIndex;not null" json:"name" yaml:"name"`
	Source       string `json:"source" yaml:"source"`
	ExecutorType string `gorm:"not null;default:'terraform-module'" json:"executorType" yaml:"executorType"`
}

type ModuleVersion struct {
	ID        uint           `gorm:"primarykey" json:"id" yaml:"id"`
	Module    Module         `gorm:"foreignKey:ModuleID" json:"module" yaml:"module"`
	ModuleID  uint           `json:"moduleId" yaml:"moduleId"`
	Version   string         `json:"version" yaml:"version"`
	Variables datatypes.JSON `gorm:"type:jsonb" json:"variables" yaml:"variables"`
	Outputs   datatypes.JSON `gorm:"type:jsonb" json:"outputs" yaml:"outputs"`
}

type GetModuleRequest struct {
	ModuleID uint `json:"moduleId" yaml:"moduleId"`
}

type GetModuleResponse struct {
	Module        Module         `json:"module" yaml:"module"`
	LatestVersion *ModuleVersion `json:"latestVersion,omitempty" yaml:"latestVersion,omitempty"`
}

type ListModulesRequest struct{}

type ListModulesResponse struct {
	Modules []Module `json:"modules" yaml:"modules"`
}

type CreateModuleRequest struct {
	Name         string `json:"name" yaml:"name"`
	Source       string `json:"source" yaml:"source"`
	Version      string `json:"version" yaml:"version"`
	ExecutorType string `json:"executorType,omitempty" yaml:"executorType,omitempty"`
}

type CreateModuleResponse struct {
	ID        uint   `json:"id" yaml:"id"`
	Name      string `json:"name" yaml:"name"`
	Source    string `json:"source" yaml:"source"`
	VersionID uint   `json:"versionId" yaml:"versionId"`
	Version   string `json:"version" yaml:"version"`
}

type UpdateModuleRequest struct {
	ModuleID uint   `json:"moduleId" yaml:"moduleId"`
	Version  string `json:"version" yaml:"version"`
}

type UpdateModuleResponse struct {
	ModuleID  uint   `json:"moduleId" yaml:"moduleId"`
	VersionID uint   `json:"versionId" yaml:"versionId"`
	Version   string `json:"version" yaml:"version"`
}

type DeleteModuleRequest struct {
	ModuleID uint `json:"moduleId" yaml:"moduleId"`
}

type DeleteModuleResponse struct {
	ModuleID uint `json:"moduleId" yaml:"moduleId"`
}

type GetModuleVersionRequest struct {
	ModuleVersionID uint `json:"moduleVersionId" yaml:"moduleVersionId"`
}

type GetModuleVersionResponse struct {
	ModuleVersion ModuleVersion `json:"moduleVersion" yaml:"moduleVersion"`
}

type ListModuleVersionsRequest struct {
	ModuleID *uint `json:"moduleId,omitempty" yaml:"moduleId,omitempty"`
}

type ListModuleVersionsResponse struct {
	ModuleVersions []ModuleVersion `json:"moduleVersions" yaml:"moduleVersions"`
}
