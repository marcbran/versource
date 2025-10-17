package versource

type Rebase struct {
	ID          uint      `gorm:"primarykey" json:"id" yaml:"id"`
	Changeset   Changeset `gorm:"foreignKey:ChangesetID" json:"changeset" yaml:"changeset"`
	ChangesetID uint      `json:"changesetId" yaml:"changesetId"`
	MergeBase   string    `gorm:"column:merge_base" json:"mergeBase" yaml:"mergeBase"`
	Head        string    `gorm:"column:head" json:"head" yaml:"head"`
	State       TaskState `gorm:"default:Queued" json:"state" yaml:"state"`
}

type GetRebaseRequest struct {
	RebaseID      uint   `json:"rebaseId" yaml:"rebaseId"`
	ChangesetName string `json:"changesetName" yaml:"changesetName"`
}

type GetRebaseResponse struct {
	Rebase Rebase `json:"rebase" yaml:"rebase"`
}

type ListRebasesRequest struct {
	ChangesetName string `json:"changesetName" yaml:"changesetName"`
}

type ListRebasesResponse struct {
	Rebases []Rebase `json:"rebases" yaml:"rebases"`
}

type CreateRebaseRequest struct {
	ChangesetName string `json:"changesetName" yaml:"changesetName"`
}

type CreateRebaseResponse struct {
	Rebase Rebase `json:"rebase" yaml:"rebase"`
}
