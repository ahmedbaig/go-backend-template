package events

import (
	"strconv"
	"strings"

	"github.com/ahmedbaig/go-backend-http"
	"github.com/ahmedbaig/go-backend-types"
)

func NewTopics(ts ...Topic) Topics {
	m := Topics{}
	for _, t := range ts {
		m[t.Name] = t
	}
	return m
}

func UserChannel(topic string, userID int64) string {
	return topic + ":" + strconv.FormatInt(userID, 10)
}

func (ts Topics) Channel(topic string, u types.User) string {
	if t, ok := ts[topic]; ok && t.PerUser {
		return UserChannel(topic, u.ID)
	}
	return topic
}

func (ts Topics) Resolve(u types.User, raw string) ([]string, error) {
	var out []string
	seen := map[string]bool{}
	for _, name := range strings.Split(raw, ",") {
		name = strings.TrimSpace(name)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		t, ok := ts[name]
		if !ok {
			return nil, httpx.E(httpx.CodeValidation).WithFields(httpx.FieldError{Field: "topics", Rule: "oneof", Message: "unknown topic " + name})
		}
		if t.Scope != "" && !u.Can(t.Scope) {
			return nil, httpx.E(httpx.CodeForbidden).WithMessage("You do not have access to " + name)
		}
		out = append(out, ts.Channel(name, u))
	}
	if len(out) == 0 {
		return nil, httpx.E(httpx.CodeValidation).WithFields(httpx.FieldError{Field: "topics", Rule: "required", Message: "at least one topic is required"})
	}
	return out, nil
}
