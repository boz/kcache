package kcache

import (
	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/filter"
)

func SubscribeWithFilter(log logutil.Log, publisher Publisher, filter filter.Filter) FilterSubscription {
	parent := publisher.Subscribe()
	return NewFilterSubscription(log, parent, filter)
}

func Clone(log logutil.Log, publisher Publisher) Controller {
	parent := publisher.Subscribe()
	return NewPublisher(log, parent)
}

func CloneWithFilter(log logutil.Log, publisher Publisher, filter filter.Filter) FilterController {
	parent := SubscribeWithFilter(log, publisher, filter)
	return NewFilterPublisher(log, parent)
}
