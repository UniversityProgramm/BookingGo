package enum

type OutboxStatus string

const (
	StatusPending OutboxStatus = "pending"
	StatusSent    OutboxStatus = "sent"
	StatusFailed  OutboxStatus = "failed"
)
