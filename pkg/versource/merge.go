package versource

type Merge struct {
	ID          uint      `gorm:"primarykey" json:"id" yaml:"id"`
	Changeset   Changeset `gorm:"foreignKey:ChangesetID" json:"changeset" yaml:"changeset"`
	ChangesetID uint      `json:"changesetId" yaml:"changesetId"`
	MergeBase   string    `gorm:"column:merge_base" json:"mergeBase" yaml:"mergeBase"`
	Head        string    `gorm:"column:head" json:"head" yaml:"head"`
	State       TaskState `gorm:"default:Queued" json:"state" yaml:"state"`
}

type GetMergeRequest struct {
	MergeID       uint   `json:"mergeId" yaml:"mergeId"`
	ChangesetName string `json:"changesetName" yaml:"changesetName"`
}

type GetMergeResponse struct {
	Merge Merge `json:"merge" yaml:"merge"`
}

type ListMergesRequest struct {
	ChangesetName string `json:"changesetName" yaml:"changesetName"`
}

type ListMergesResponse struct {
	Merges []Merge `json:"merges" yaml:"merges"`
}

type CreateMergeRequest struct {
	ChangesetName string `json:"changesetName" yaml:"changesetName"`
}

type CreateMergeResponse struct {
	Merge Merge `json:"merge" yaml:"merge"`
}
