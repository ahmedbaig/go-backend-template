package events

import (
	"context"
	"log/slog"
	"strconv"
	"strings"

	"github.com/ahmedbaig/go-backend-http"
	"github.com/ahmedbaig/go-backend-http/sse"
	"github.com/ahmedbaig/go-backend-proto/eventsx"
	eventsv1 "github.com/ahmedbaig/go-backend-proto/gen/go/ahmedbaig/events/v1"
	"github.com/ahmedbaig/go-backend-template/internal/constants"
	"github.com/redis/go-redis/v9"
)

func NewBroker(rdb redis.UniversalClient, o Options) *Broker {
	if o.MaxLen <= 0 {
		o.MaxLen = 1000
	}
	return &Broker{rdb: rdb, o: o}
}

func (b *Broker) stream(channel string) string {
	return b.o.Prefix + constants.KeyEventsStream + channel
}
func (b *Broker) channel(channel string) string {
	return b.o.Prefix + constants.KeyEventsChannel + channel
}

func (b *Broker) Publish(ctx context.Context, channel, typ string, data any) (string, error) {
	ev, err := eventsx.New(ctx, strings.SplitN(channel, ":", 2)[0], typ, data)
	if err != nil {
		return "", err
	}
	body, err := eventsx.Marshal(ev)
	if err != nil {
		return "", err
	}
	id, err := b.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: b.stream(channel),
		MaxLen: b.o.MaxLen,
		Approx: true,
		Values: map[string]any{"e": body},
	}).Result()
	if err != nil {
		return "", err
	}
	ev.Id = id
	if body, err = eventsx.Marshal(ev); err != nil {
		return id, err
	}
	return id, b.rdb.Publish(ctx, b.channel(channel), body).Err()
}

func (b *Broker) PublishOrLog(ctx context.Context, channel, typ string, data any) {
	if _, err := b.Publish(ctx, channel, typ, data); err != nil {
		httpx.LoggerFrom(ctx).WarnContext(ctx, "event not published", slog.String("channel", channel), slog.String("type", typ), slog.Any("error", err))
	}
}

func (b *Broker) Subscribe(ctx context.Context, channels []string, lastID string) (<-chan sse.Event, error) {
	names := make([]string, len(channels))
	for i, c := range channels {
		names[i] = b.channel(c)
	}
	ps := b.rdb.Subscribe(ctx, names...)
	if _, err := ps.Receive(ctx); err != nil {
		_ = ps.Close()
		return nil, err
	}
	out := make(chan sse.Event, 64)
	go func() {
		defer close(out)
		defer ps.Close()
		seen := map[string]string{}
		if lastID != "" {
			for _, c := range channels {
				msgs, err := b.rdb.XRange(ctx, b.stream(c), "("+lastID, "+").Result()
				if err != nil {
					continue
				}
				for _, m := range msgs {
					raw, _ := m.Values["e"].(string)
					ev, err := eventsx.Unmarshal([]byte(raw))
					if err != nil {
						continue
					}
					ev.Id = m.ID
					if !send(ctx, out, ev) {
						return
					}
					seen[ev.GetTopic()] = m.ID
				}
			}
		}
		msgs := ps.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case m, ok := <-msgs:
				if !ok {
					return
				}
				ev, err := eventsx.Unmarshal([]byte(m.Payload))
				if err != nil {
					continue
				}
				if last, ok := seen[ev.GetTopic()]; ok && !After(ev.GetId(), last) {
					continue
				}
				if !send(ctx, out, ev) {
					return
				}
			}
		}
	}()
	return out, nil
}

func send(ctx context.Context, out chan<- sse.Event, ev *eventsv1.Event) bool {
	body, err := eventsx.Marshal(ev)
	if err != nil {
		return true
	}
	select {
	case out <- sse.Event{ID: ev.GetId(), Type: ev.GetType(), Data: body}:
		return true
	case <-ctx.Done():
		return false
	}
}

func After(a, b string) bool {
	am, as := split(a)
	bm, bs := split(b)
	if am != bm {
		return am > bm
	}
	return as > bs
}

func split(id string) (uint64, uint64) {
	ms, seq, _ := strings.Cut(id, "-")
	m, _ := strconv.ParseUint(ms, 10, 64)
	s, _ := strconv.ParseUint(seq, 10, 64)
	return m, s
}
