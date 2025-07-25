package telegram2

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"read_adviser_bot/internal/clients/telegram"
	"read_adviser_bot/internal/events"
	"read_adviser_bot/internal/storage"
	"time"
)

var (
	ErrUnknownEventType = errors.New("event type is unknown")
	ErrUnknownMetaType  = errors.New("meta type is unknown")
)

type Processor struct {
	tg      *telegram.Client
	offset  int
	storage storage.LinkRepository
}

type Meta struct {
	ChatID   int
	Username string
}

func New(tg *telegram.Client, storage storage.LinkRepository) *Processor {
	return &Processor{
		tg:      tg,
		storage: storage,
	}
}

func (p *Processor) Fetch(limit int) ([]events.Event, error) {
	const op = "events.telegram.Fetch"

	updates, err := p.tg.Updates(p.offset, limit)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to get update: %w", op, err)
	}

	if len(updates) == 0 {
		return nil, nil
	}

	res := make([]events.Event, 0, len(updates))

	for _, i := range updates {
		res = append(res, event(i))
	}

	p.offset = updates[len(updates)-1].ID + 1

	return res, nil
}

func (p *Processor) Process(log *slog.Logger, event events.Event) error {
	switch event.Type {
	case events.Message:
		return p.processMessage(log, event)
	default:
		return ErrUnknownEventType
	}
}

func event(upd telegram.Update) events.Event {
	updType := fetchType(upd)

	res := events.Event{
		Type: updType,
		Text: fetchText(upd),
	}

	if updType == events.Message {
		res.Meta = Meta{
			ChatID:   upd.Message.Chat.ID,
			Username: upd.Message.From.Username,
		}
	}

	return res
}

func fetchType(upd telegram.Update) events.Type {
	if upd.Message == nil {
		return events.Unknown
	}

	return events.Message
}

func (p *Processor) processMessage(log *slog.Logger, event events.Event) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	meta, err := meta(event)
	if err != nil {
		return fmt.Errorf("can't process message %w", err)
	}

	if err := p.doCmd(ctx, log, event.Text, meta.ChatID, meta.Username); err != nil {
		return fmt.Errorf("can't process message %w", err)
	}

	return nil
}

func meta(event events.Event) (Meta, error) {
	res, ok := event.Meta.(Meta)
	if !ok {
		return Meta{}, fmt.Errorf("failed to get meta: %w", ErrUnknownMetaType)
	}

	return res, nil
}

func fetchText(upd telegram.Update) string {
	if upd.Message == nil {
		return ""
	}

	return upd.Message.Text
}
