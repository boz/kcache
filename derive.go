package kcache

import logutil "github.com/boz/go-logutil"

func SubscribeWithFilter(log logutil.Log, publisher Publisher, filter Filter) Subscription {
	parent := publisher.Subscribe()
	return NewFilterSubscription(log, parent, filter)
}

func Clone(log logutil.Log, publisher Publisher) Controller {
	parent := publisher.Subscribe()
	return NewPublisher(log, parent)
}

func CloneWithFilter(log logutil.Log, publisher Publisher, filter Filter) Controller {
	parent := SubscribeWithFilter(log, publisher, filter)
	return NewPublisher(log, parent)
}
