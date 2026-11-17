package events

import "github.com/redis/go-redis/v9"

type Options struct {
	Prefix string
	MaxLen int64
}

type Broker struct {
	rdb redis.UniversalClient
	o   Options
}

type Topic struct {
	Name    string
	Scope   string
	PerUser bool
}

type Topics map[string]Topic
