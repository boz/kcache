package event_test

import (
	"testing"

	"github.com/boz/kcache/types/event"
	"github.com/stretchr/testify/assert"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/pkg/api/v1"
)

func TestInvolvedForFilter(t *testing.T) {

	genpod := func(ns, name string) *v1.Pod {
		return &v1.Pod{
			TypeMeta:   metav1.TypeMeta{Kind: "pod"},
			ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name},
		}
	}

	genevt := func(kind, ns, name string) *v1.Event {
		return &v1.Event{
			InvolvedObject: v1.ObjectReference{
				Kind:      kind,
				Namespace: ns,
				Name:      name,
			},
		}
	}

	{
		f := event.InvolvedFilter("pod", "a", "b")
		assert.True(t, f.Accept(genevt("pod", "a", "b")))
		assert.False(t, f.Accept(genevt("pod", "a", "c")))
		assert.False(t, f.Accept(genevt("pod", "c", "b")))
		assert.False(t, f.Accept(genevt("pod", "c", "c")))
		assert.False(t, f.Accept(genevt("service", "a", "b")))
	}

	{
		f := event.InvolvedFilter("pod", "a", "b")
		assert.True(t, f.Equals(event.InvolvedFilter("pod", "a", "b")))
		assert.False(t, f.Equals(event.InvolvedFilter("pod", "a", "c")))
		assert.False(t, f.Equals(event.InvolvedFilter("pod", "c", "b")))
		assert.False(t, f.Equals(event.InvolvedFilter("pod", "c", "c")))
		assert.False(t, f.Equals(event.InvolvedFilter("service", "a", "b")))
	}

	{
		f := event.InvolvedObjectFilter(genpod("a", "b"))
		assert.True(t, f.Accept(genevt("pod", "a", "b")))
		assert.False(t, f.Accept(genevt("pod", "a", "c")))
		assert.False(t, f.Accept(genevt("pod", "c", "b")))
		assert.False(t, f.Accept(genevt("pod", "c", "c")))
		assert.False(t, f.Accept(genevt("service", "a", "b")))

		assert.True(t, f.Equals(event.InvolvedFilter("pod", "a", "b")))
		assert.False(t, f.Equals(event.InvolvedFilter("pod", "a", "c")))
		assert.False(t, f.Equals(event.InvolvedFilter("pod", "c", "b")))
		assert.False(t, f.Equals(event.InvolvedFilter("pod", "c", "c")))
		assert.False(t, f.Equals(event.InvolvedFilter("service", "a", "b")))
	}
}
