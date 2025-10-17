package versource

import (
	"io"
)

type Plan struct {
	ID          uint      `gorm:"primarykey" json:"id" yaml:"id"`
	Changeset   Changeset `gorm:"foreignKey:ChangesetID" json:"changeset" yaml:"changeset"`
	ChangesetID uint      `json:"changesetId" yaml:"changesetId"`
	ComponentID uint      `json:"componentId" yaml:"componentId"`
	From        string    `gorm:"column:from" json:"from" yaml:"from"`
	To          string    `gorm:"column:to" json:"to" yaml:"to"`
	State       TaskState `gorm:"default:Queued" json:"state" yaml:"state"`
	Add         *int      `gorm:"column:add" json:"add" yaml:"add"`
	Change      *int      `gorm:"column:change" json:"change" yaml:"change"`
	Destroy     *int      `gorm:"column:destroy" json:"destroy" yaml:"destroy"`
}

type GetPlanRequest struct {
	ChangesetName *string `json:"changesetName" yaml:"changesetName"`
	PlanID        uint    `json:"planId" yaml:"planId"`
}

type GetPlanResponse struct {
	Plan      Plan      `json:"plan" yaml:"plan"`
	Component Component `json:"component" yaml:"component"`
}

type GetPlanLogRequest struct {
	ChangesetName *string `json:"changesetName" yaml:"changesetName"`
	PlanID        uint    `json:"planId" yaml:"planId"`
}

type GetPlanLogResponse struct {
	Content io.ReadCloser `json:"content" yaml:"content"`
}

type ListPlansRequest struct {
	ChangesetName string `json:"changesetName" yaml:"changesetName"`
}

type ListPlansResponse struct {
	Plans []Plan `json:"plans" yaml:"plans"`
}

type CreatePlanRequest struct {
	ComponentID   uint   `json:"componentId" yaml:"componentId"`
	ChangesetName string `json:"changesetName" yaml:"changesetName"`
}

type CreatePlanResponse struct {
	Plan Plan `json:"plan" yaml:"plan"`
}
