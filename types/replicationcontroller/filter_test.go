package replicationcontroller_test

import (
	"testing"

	"github.com/boz/kcache/types/replicationcontroller"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPodsFilter(t *testing.T) {

	genpod := func(labels map[string]string) *v1.Pod {
		return &v1.Pod{ObjectMeta: metav1.ObjectMeta{Labels: labels}}
	}

	gensvc := func(ns, name string, labels map[string]string) *v1.ReplicationController {
		return &v1.ReplicationController{
			ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name},
			Spec:       v1.ReplicationControllerSpec{Selector: labels},
		}
	}

	p1 := genpod(map[string]string{"a": "1", "b": "1", "c": "x"})
	p2 := genpod(map[string]string{"a": "2", "b": "2", "c": "x"})

	s1 := gensvc("a", "1", map[string]string{"a": "1"})
	s2 := gensvc("a", "2", map[string]string{"b": "2"})
	s3 := gensvc("c", "1", map[string]string{"c": "x"})
	s4 := gensvc("d", "1", map[string]string{"a": "0"})

	assert.False(t, replicationcontroller.PodsFilter().Accept(p1))
	assert.True(t, replicationcontroller.PodsFilter(s1).Accept(p1))
	assert.False(t, replicationcontroller.PodsFilter(s1).Accept(p2))
	assert.True(t, replicationcontroller.PodsFilter(s2).Accept(p2))
	assert.False(t, replicationcontroller.PodsFilter(s2).Accept(p1))

	assert.True(t, replicationcontroller.PodsFilter(s1, s2).Accept(p1))

	assert.True(t, replicationcontroller.PodsFilter(s3).Accept(p1))
	assert.True(t, replicationcontroller.PodsFilter(s3).Accept(p2))

	assert.False(t, replicationcontroller.PodsFilter(s4).Accept(p1))
	assert.False(t, replicationcontroller.PodsFilter(s4).Accept(p2))

	assert.True(t, replicationcontroller.PodsFilter(s1).Equals(replicationcontroller.PodsFilter(s1)))
	assert.True(t, replicationcontroller.PodsFilter(s1, s2).Equals(replicationcontroller.PodsFilter(s1, s2)))
	assert.True(t, replicationcontroller.PodsFilter(s2, s1).Equals(replicationcontroller.PodsFilter(s1, s2)))
	assert.True(t, replicationcontroller.PodsFilter(s4, s3).Equals(replicationcontroller.PodsFilter(s3, s4)))

}
