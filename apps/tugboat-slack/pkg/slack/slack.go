package slack

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/object88/tugboat/apps/tugboat-slack/pkg/slack/config"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type Bot struct {
	Logger *logrus.Logger

	api *slack.Client
	cfg *config.Config
}

func New(cfg *config.Config) *Bot {
	api := slack.New(cfg.Token)

	return &Bot{
		api: api,
		cfg: cfg,
	}
}

func (b *Bot) PreprocessSecurity(req *http.Request) (*slack.SecretsVerifier, error) {
	sv, err := slack.NewSecretsVerifier(req.Header, b.cfg.SigningSecret)
	if err != nil {
		// TODO: handle this error.
	}
	req.Body = ioutil.NopCloser(io.TeeReader(req.Body, &sv))

	return &sv, nil
}

func (b *Bot) ProcessSecurity(sv *slack.SecretsVerifier) error {
	return sv.Ensure()
}

func (b *Bot) ProcessEventCommand(w http.ResponseWriter, r *http.Request) {
	sv, err := b.PreprocessSecurity(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Do some things.
	var buf bytes.Buffer
	buf.ReadFrom(r.Body)
	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(buf.Bytes()), slackevents.OptionVerifyToken(&slackevents.TokenComparator{VerificationToken: b.cfg.Verification}))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := b.ProcessSecurity(sv); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if eventsAPIEvent.Type == slackevents.URLVerification {
		r, ok := eventsAPIEvent.Data.(*slackevents.EventsAPIURLVerificationEvent)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.Header().Set("Content-Type", "text")
			w.Write([]byte(r.Challenge))
		}
	} else if eventsAPIEvent.Type == slackevents.CallbackEvent {
		innerEvent := eventsAPIEvent.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			b.SendMessage(ev.Channel, "Yes, hello.")
		}
	}
}

func (b *Bot) ProcessSlashCommand(w http.ResponseWriter, r *http.Request) {
	sv, err := slack.NewSecretsVerifier(r.Header, b.cfg.SigningSecret)
	r.Body = ioutil.NopCloser(io.TeeReader(r.Body, &sv))

	s, err := slack.SlashCommandParse(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err = sv.Ensure(); err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	b.api.PostMessage(s.ChannelID, slack.MsgOptionText(s.Text, false))

	w.WriteHeader(200)
}

func (b *Bot) ProcessInteractiveCommand(w http.ResponseWriter, r *http.Request) {
}

func (b *Bot) SendMessage(channel string, msg string) error {
	_, _, err := b.api.PostMessage(channel, slack.MsgOptionText(msg, false))
	return err
}
