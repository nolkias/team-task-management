package domain

import "time"

type TaskHistory struct {
	ID           int64
	TaskID       int64
	ChangedBy    int64
	FieldChanged string
	OldValue     string
	NewValue     string
	ChangedAt    time.Time
}
