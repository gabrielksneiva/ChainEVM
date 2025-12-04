package events

import "time"

// DomainEvent interface para todos os eventos de domínio
type DomainEvent interface {
	EventType() string
	AggregateID() string
	OccurredAt() time.Time
}

// BaseDomainEvent implementação base para eventos de domínio
type BaseDomainEvent struct {
	eventType   string
	aggregateID string
	occurredAt  time.Time
}

// NewBaseDomainEvent cria um novo evento base
func NewBaseDomainEvent(eventType, aggregateID string) *BaseDomainEvent {
	return &BaseDomainEvent{
		eventType:   eventType,
		aggregateID: aggregateID,
		occurredAt:  time.Now(),
	}
}

// EventType retorna o tipo do evento
func (e *BaseDomainEvent) EventType() string {
	return e.eventType
}

// AggregateID retorna o ID do agregado
func (e *BaseDomainEvent) AggregateID() string {
	return e.aggregateID
}

// OccurredAt retorna quando o evento ocorreu
func (e *BaseDomainEvent) OccurredAt() time.Time {
	return e.occurredAt
}
