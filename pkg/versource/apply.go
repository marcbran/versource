package versource

import (
	"io"
)

type TaskState string

const (
	TaskStateQueued    TaskState = "Queued"
	TaskStateStarted   TaskState = "Started"
	TaskStateAborted   TaskState = "Aborted"
	TaskStateSucceeded TaskState = "Succeeded"
	TaskStateFailed    TaskState = "Failed"
	TaskStateCancelled TaskState = "Cancelled"
)

func IsTaskCompleted(task TaskState) bool {
	return task == TaskStateSucceeded || task == TaskStateFailed || task == TaskStateCancelled
}

type Apply struct {
	ID          uint      `gorm:"primarykey" json:"id" yaml:"id"`
	Plan        Plan      `gorm:"foreignKey:PlanID" json:"plan" yaml:"plan"`
	PlanID      uint      `gorm:"uniqueIndex" json:"planId" yaml:"planId"`
	Changeset   Changeset `gorm:"foreignKey:ChangesetID" json:"changeset" yaml:"changeset"`
	ChangesetID uint      `json:"changesetId" yaml:"changesetId"`
	State       TaskState `gorm:"default:Queued" json:"state" yaml:"state"`
}

type GetApplyRequest struct {
	ApplyID uint `json:"applyId" yaml:"applyId"`
}

type GetApplyResponse struct {
	Apply     Apply     `json:"apply" yaml:"apply"`
	Component Component `json:"component" yaml:"component"`
}

type GetApplyLogRequest struct {
	ApplyID uint `json:"applyId" yaml:"applyId"`
}

type GetApplyLogResponse struct {
	Content io.ReadCloser `json:"content" yaml:"content"`
}

type ListAppliesRequest struct{}

type ListAppliesResponse struct {
	Applies []Apply `json:"applies" yaml:"applies"`
}
