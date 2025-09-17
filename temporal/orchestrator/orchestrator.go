package orchestrator

import "errors"

type Orchestrator interface {
	RegisterItem(itemID string, itemWorkflowID string, itemWorkflowRunID string, item interface{}) error
	StartProcessing(itemID string) (*OrchestratedItem, error)
	StopProcessing(itemID string) (*OrchestratedItem, error)
	UpdateItem(itemID string, item interface{}) error
	Deregister(itemID string) error
	AllItems() map[string]OrchestratedItem
	RegisteredItems() map[string]OrchestratedItem
}

type OrchestratedItem struct {
	ID                string      `json:"id"`
	ItemWorkflowID    string      `json:"itemWorkflowId"`
	ItemWorkflowRunID string      `json:"itemWorkflowRunId"`
	InProgress        bool        `json:"inProgress"`
	Deregistered      bool        `json:"deregistered"`
	Payload           interface{} `json:"payload"`
}

type ItemOrchestrator struct {
	Items map[string]OrchestratedItem
}

var _ Orchestrator = (*ItemOrchestrator)(nil)

var (
	itemNotRegisteredError = errors.New("item not registered")
	anotherItemInProgress  = errors.New("another item already in progress")
)

func NewItemOrchestrator() ItemOrchestrator {
	return ItemOrchestrator{
		Items: make(map[string]OrchestratedItem, 0),
	}
}

func (o *ItemOrchestrator) RegisterItem(itemID string, itemWorkflowID string, itemWorkflowRunID string, item interface{}) error {
	newItem := OrchestratedItem{
		ID:                itemID,
		ItemWorkflowID:    itemWorkflowID,
		ItemWorkflowRunID: itemWorkflowRunID,
		Payload:           item,
	}
	if itemInProgress := o.getItemInProgress(); itemInProgress != nil && itemInProgress.ID != itemID {
		newItem.Deregistered = true
		o.Items[itemID] = newItem
		return anotherItemInProgress
	}
	o.Items[itemID] = newItem
	return nil
}

func (o *ItemOrchestrator) StartProcessing(itemID string) (*OrchestratedItem, error) {
	if item, exists := o.Items[itemID]; exists {
		if itemInProgress := o.getItemInProgress(); itemInProgress != nil && itemInProgress.ID != itemID {
			return &item, anotherItemInProgress
		}
		item.InProgress = true
		o.Items[itemID] = item
		return &item, nil
	} else {
		return nil, itemNotRegisteredError
	}
}

func (o *ItemOrchestrator) StopProcessing(itemID string) (*OrchestratedItem, error) {
	if item, exists := o.Items[itemID]; exists {
		item.InProgress = false
		o.Items[itemID] = item
		return &item, nil
	} else {
		return nil, itemNotRegisteredError
	}
}

func (o *ItemOrchestrator) UpdateItem(itemID string, item interface{}) error {
	if existingItem, exists := o.Items[itemID]; exists {
		existingItem.Payload = item
		o.Items[itemID] = existingItem
	} else {
		return itemNotRegisteredError
	}
	return nil
}

func (o *ItemOrchestrator) Deregister(itemID string) error {
	if item, exists := o.Items[itemID]; exists {
		item.Deregistered = true
		o.Items[itemID] = item
	} else {
		return itemNotRegisteredError
	}
	return nil
}

func (o *ItemOrchestrator) AllItems() map[string]OrchestratedItem {
	return o.Items
}

func (o *ItemOrchestrator) RegisteredItems() map[string]OrchestratedItem {
	registered := make(map[string]OrchestratedItem)
	for id, item := range o.Items {
		if !item.Deregistered {
			registered[id] = item
		}
	}
	return registered
}

func (o *ItemOrchestrator) getItemInProgress() *OrchestratedItem {
	for _, item := range o.Items {
		if item.InProgress && !item.Deregistered {
			return &item
		}
	}
	return nil
}
