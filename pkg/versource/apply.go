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
	ID          uint      `json:"id" yaml:"id"`
	PlanID      uint      `json:"planId" yaml:"planId"`
	ChangesetID uint      `json:"changesetId" yaml:"changesetId"`
	State       TaskState `json:"state" yaml:"state"`
	Plan        struct {
		ID          uint      `json:"id" yaml:"id"`
		State       TaskState `json:"state" yaml:"state"`
		From        string    `json:"from" yaml:"from"`
		To          string    `json:"to" yaml:"to"`
		Add         *int      `json:"add,omitempty" yaml:"add,omitempty"`
		Change      *int      `json:"change,omitempty" yaml:"change,omitempty"`
		Destroy     *int      `json:"destroy,omitempty" yaml:"destroy,omitempty"`
		ComponentID uint      `json:"componentId" yaml:"componentId"`
		Component   struct {
			ID   uint   `json:"id" yaml:"id"`
			Name string `json:"name" yaml:"name"`
		} `json:"component" yaml:"component"`
		ChangesetID uint `json:"changesetId" yaml:"changesetId"`
		Changeset   struct {
			ID   uint   `json:"id" yaml:"id"`
			Name string `json:"name" yaml:"name"`
		} `json:"changeset" yaml:"changeset"`
	} `json:"plan" yaml:"plan"`
	Changeset struct {
		ID   uint   `json:"id" yaml:"id"`
		Name string `json:"name" yaml:"name"`
	} `json:"changeset" yaml:"changeset"`
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
