package orchestrator

type Item interface {
	ID() string
	GetStatus() string
	GetName() string
}

type BasicItem struct {
	Id     string     `json:"id"`
	Name   string     `json:"name"`
	Status ItemStatus `json:"status"`
}

var _ Item = (*BasicItem)(nil)

func (i BasicItem) ID() string {
	return i.Id
}

func (i BasicItem) GetStatus() string {
	return i.Status.String()
}

func (i BasicItem) GetName() string {
	return i.Name
}

type ItemA struct {
	BasicItem
	ExtraFieldA string `json:"extraFieldA"`
}

var _ Item = (*ItemA)(nil)

type ItemB struct {
	BasicItem
	ExtraFieldB string `json:"extraFieldB"`
}

var _ Item = (*ItemB)(nil)

type ItemStatus string

const (
	ItemStatusNew        ItemStatus = "New"
	ItemStatusProcessing ItemStatus = "Processing"
	ItemStatusCompleted  ItemStatus = "Completed"
	ItemStatusFailed     ItemStatus = "Failed"
	ItemStatusCancelled  ItemStatus = "Cancelled"
)

func (s ItemStatus) String() string {
	return string(s)
}
