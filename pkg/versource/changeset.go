package versource

type Changeset struct {
	ID          uint                 `gorm:"primarykey" json:"id" yaml:"id"`
	Name        string               `gorm:"index" json:"name" yaml:"name"`
	State       ChangesetState       `gorm:"default:Open" json:"state" yaml:"state"`
	ReviewState ChangesetReviewState `gorm:"default:Draft" json:"reviewState" yaml:"reviewState"`
}

type ChangesetState string

const (
	ChangesetStateOpen   ChangesetState = "Open"
	ChangesetStateClosed ChangesetState = "Closed"
	ChangesetStateMerged ChangesetState = "Merged"
)

type ChangesetReviewState string

const (
	ChangesetReviewStateDraft    ChangesetReviewState = "Draft"
	ChangesetReviewStatePending  ChangesetReviewState = "Pending"
	ChangesetReviewStateApproved ChangesetReviewState = "Approved"
	ChangesetReviewStateRejected ChangesetReviewState = "Rejected"
)

type ListChangesetsRequest struct{}

type ListChangesetsResponse struct {
	Changesets []Changeset `json:"changesets" yaml:"changesets"`
}

type CreateChangesetRequest struct {
	Name string `json:"name" yaml:"name"`
}

type CreateChangesetResponse struct {
	ID    uint           `json:"id" yaml:"id"`
	Name  string         `json:"name" yaml:"name"`
	State ChangesetState `json:"state" yaml:"state"`
}

type EnsureChangesetRequest struct {
	Name string `json:"name" yaml:"name"`
}

type EnsureChangesetResponse struct {
	ID    uint           `json:"id" yaml:"id"`
	Name  string         `json:"name" yaml:"name"`
	State ChangesetState `json:"state" yaml:"state"`
}

type DeleteChangesetRequest struct {
	ChangesetName string `json:"changesetName" yaml:"changesetName"`
}

type DeleteChangesetResponse struct {
	ID uint `json:"id" yaml:"id"`
}
