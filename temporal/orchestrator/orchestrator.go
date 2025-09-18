package orchestrator

import (
	"errors"
)

type OrchestratorState struct {
	SignalsHandled    int
	OrchestratedItems map[string]OrchestratedItem
}

func (o *OrchestratorState) IncrementSignalsHandled() {
	if o != nil {
		o.SignalsHandled++
	}
}

func (o *OrchestratorState) GetSignalsHandled() int {
	if o == nil {
		return 0
	}
	return o.SignalsHandled
}

type OrchestratorStateManager interface {
	GetState() *OrchestratorState
	RegisterItem(itemID string, itemWorkflowID string, itemWorkflowRunID string, item interface{}) error
	StartProcessing(itemID string) (*OrchestratedItem, error)
	StopProcessing(itemID string) (*OrchestratedItem, error)
	UpdateItem(itemID string, item interface{}) error
	Deregister(itemID string) error
	AllItems() map[string]OrchestratedItem
	RegisteredItems() map[string]OrchestratedItem
}

type CreateOrchestratorStateManagerFunc[O OrchestratorStateManager] func(state *OrchestratorState) O

type OrchestratedItem struct {
	ID                string      `json:"id"`
	ItemWorkflowID    string      `json:"itemWorkflowId"`
	ItemWorkflowRunID string      `json:"itemWorkflowRunId"`
	InProgress        bool        `json:"inProgress"`
	Deregistered      bool        `json:"deregistered"`
	Payload           interface{} `json:"payload"`
}

type ItemOrchestratorStateManager struct {
	state *OrchestratorState
}

var _ OrchestratorStateManager = (*ItemOrchestratorStateManager)(nil)

var (
	itemNotRegisteredError = errors.New("item not registered")
	anotherItemInProgress  = errors.New("another item already in progress")
)

func NewItemOrchestratorStateManager(state *OrchestratorState) OrchestratorStateManager {
	osm := &ItemOrchestratorStateManager{}
	if state == nil {
		osm.state = &OrchestratorState{}
	} else {
		osm.state = state
	}
	if osm.state.OrchestratedItems == nil {
		osm.state.OrchestratedItems = make(map[string]OrchestratedItem)
	}
	return osm
}

func (o *ItemOrchestratorStateManager) GetState() *OrchestratorState {
	if o == nil {
		return nil
	}
	return o.state
}

func (o *ItemOrchestratorStateManager) RegisterItem(itemID string, itemWorkflowID string, itemWorkflowRunID string, item interface{}) error {
	if o == nil {
		return errors.New("orchestrator state manager is nil")
	}
	newItem := OrchestratedItem{
		ID:                itemID,
		ItemWorkflowID:    itemWorkflowID,
		ItemWorkflowRunID: itemWorkflowRunID,
		Payload:           item,
	}
	if itemInProgress := o.getItemInProgress(); itemInProgress != nil && itemInProgress.ID != itemID {
		newItem.Deregistered = true
		o.state.OrchestratedItems[itemID] = newItem
		return anotherItemInProgress
	}
	o.state.OrchestratedItems[itemID] = newItem
	return nil
}

func (o *ItemOrchestratorStateManager) StartProcessing(itemID string) (*OrchestratedItem, error) {
	if o == nil {
		return nil, errors.New("orchestrator state manager is nil")
	}
	if item, exists := o.state.OrchestratedItems[itemID]; exists {
		if itemInProgress := o.getItemInProgress(); itemInProgress != nil && itemInProgress.ID != itemID {
			return &item, anotherItemInProgress
		}
		item.InProgress = true
		o.state.OrchestratedItems[itemID] = item
		return &item, nil
	} else {
		return nil, itemNotRegisteredError
	}
}

func (o *ItemOrchestratorStateManager) StopProcessing(itemID string) (*OrchestratedItem, error) {
	if o == nil {
		return nil, errors.New("orchestrator state manager is nil")
	}
	if item, exists := o.state.OrchestratedItems[itemID]; exists {
		item.InProgress = false
		o.state.OrchestratedItems[itemID] = item
		return &item, nil
	} else {
		return nil, itemNotRegisteredError
	}
}

func (o *ItemOrchestratorStateManager) UpdateItem(itemID string, item interface{}) error {
	if o == nil {
		return errors.New("orchestrator state manager is nil")
	}
	if existingItem, exists := o.state.OrchestratedItems[itemID]; exists {
		existingItem.Payload = item
		o.state.OrchestratedItems[itemID] = existingItem
	} else {
		return itemNotRegisteredError
	}
	return nil
}

func (o *ItemOrchestratorStateManager) Deregister(itemID string) error {
	if o == nil {
		return errors.New("orchestrator state manager is nil")
	}
	if item, exists := o.state.OrchestratedItems[itemID]; exists {
		item.Deregistered = true
		o.state.OrchestratedItems[itemID] = item
	} else {
		return itemNotRegisteredError
	}
	return nil
}

func (o *ItemOrchestratorStateManager) AllItems() map[string]OrchestratedItem {
	if o == nil {
		return nil
	}
	return o.state.OrchestratedItems
}

func (o *ItemOrchestratorStateManager) RegisteredItems() map[string]OrchestratedItem {
	if o == nil {
		return nil
	}
	registered := make(map[string]OrchestratedItem)
	for id, item := range o.state.OrchestratedItems {
		if !item.Deregistered {
			registered[id] = item
		}
	}
	return registered
}

func (o *ItemOrchestratorStateManager) getItemInProgress() *OrchestratedItem {
	if o == nil {
		return nil
	}
	for _, item := range o.state.OrchestratedItems {
		if item.InProgress && !item.Deregistered {
			return &item
		}
	}
	return nil
}
