package event_consumer

import (
	"log/slog"
	"time"

	"read_adviser_bot/internal/events"
	"read_adviser_bot/internal/lib/logger/sl"
)

type Consumer struct {
	fetcher   events.Fetcher
	processor events.Processor
	batchSize int
}

func New(fetcher events.Fetcher, processor events.Processor, batchSize int) Consumer {
	return Consumer{
		fetcher:   fetcher,
		processor: processor,
		batchSize: batchSize,
	}
}

func (c Consumer) Start(log *slog.Logger) error {
	for {
		gotEvents, err := c.fetcher.Fetch(c.batchSize)
		if err != nil {
			log.Error("consumer error", sl.Err(err))

			continue
		}

		if len(gotEvents) == 0 {
			time.Sleep(1 * time.Second)

			continue
		}

		if err := c.handleEvents(gotEvents, log); err != nil {
			continue
		}
	}
}

func (c *Consumer) handleEvents(events []events.Event, log *slog.Logger) error {
	for _, event := range events {
		log.Info("got new event", slog.String("event text", event.Text))

		if err := c.processor.Process(log, event); err != nil {
			log.Error("failed to process event", sl.Err(err))

			continue
		}
	}

	return nil
}
