package telegram2

import (
	"context"
	"errors"
	"log/slog"
	"read_adviser_bot/internal/lib/logger/sl"
	"read_adviser_bot/internal/storage"
	"regexp"
	"strconv"
	"strings"
)

const (
	CmdRand  = "/rnd"
	CmdHelp  = "/help"
	CmdStart = "/start"
)

var (
	urlValidation = regexp.MustCompile(
		`(?i)^https?://[^\s/$.?#].[^\s]*$`,
	)
)

func (p *Processor) doCmd(ctx context.Context, log *slog.Logger, text string, chatID int, username string) error {
	text = strings.TrimSpace(text)

	if isAddCmd(text) {
		return p.savePage(ctx, log, text, chatID, username)
	}

	switch text {
	case CmdHelp:
		return p.sendHelp(chatID)
	case CmdStart:
		return p.sendHello(chatID)
	case CmdRand:
		return p.sendRandom(ctx, log, chatID, username)
	default:
		return p.unknownCommand(chatID)
	}
}

func isAddCmd(text string) bool {
	return isURL(text)
}

func isURL(text string) bool {
	return urlValidation.MatchString(text)
}

func (p *Processor) sendHelp(chatID int) error {
	return p.tg.SendMessage(chatID, msgHelp)
}

func (p *Processor) sendHello(chatID int) error {
	return p.tg.SendMessage(chatID, msgHello)
}

func (p *Processor) unknownCommand(chatID int) error {
	return p.tg.SendMessage(chatID, msgUnknownCommand)
}

func (p *Processor) savePage(ctx context.Context, log *slog.Logger, text string, chatID int, username string) error {
	id, err := p.storage.SaveLink(ctx, text, username)
	if err != nil {
		if errors.Is(err, storage.ErrAlreadyExists) {
			log.Warn("link is already exists")

			return p.tg.SendMessage(chatID, msgLinkExists)
		}

		log.Error("failed to add link", sl.Err(err))

		return err
	}

	msg := msgLinkSaved + strconv.FormatInt(id, 10)

	if err := p.tg.SendMessage(chatID, msg); err != nil {
		log.Error("failed to send message", sl.Err(err))

		return err
	}

	log.Info("link successfully saved",
		slog.String("url", text),
		slog.String("user", username),
		slog.Int("chat_id", chatID),
	)

	return nil
}

func (p *Processor) sendRandom(ctx context.Context, log *slog.Logger, chatID int, username string) error {
	var l *storage.Link

	l, err := p.storage.RandomLink(ctx, username)
	if err != nil {
		if errors.Is(err, storage.ErrEmptyTable) {
			log.Warn("table is empty")

			return p.tg.SendMessage(chatID, msgNoSavedLinks)
		}

		log.Error("failed to get link", sl.Err(err))

		return err
	}

	if err := p.tg.SendMessage(chatID, l.Link); err != nil {
		log.Error("failed to send message", sl.Err(err))

		return err
	}

	return nil
}
