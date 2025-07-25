package events

import "log/slog"

type Type int

const (
	Unknown Type = iota
	Message
)

type Fetcher interface {
	Fetch(limit int) ([]Event, error)
}

type Processor interface {
	Process(log *slog.Logger, e Event) error
}

type Event struct {
	Type Type
	Text string
	Meta interface{}
}
