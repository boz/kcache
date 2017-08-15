package replicaset_test

import (
	"testing"

	"github.com/boz/kcache/types/replicaset"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPodsFilter_selector(t *testing.T) {

	genselector := func(ns, name string, labels map[string]string) *v1beta1.ReplicaSet {
		return &v1beta1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name},
			Spec: v1beta1.ReplicaSetSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: labels,
				},
			},
		}
	}

	gentemplate := func(ns, name string, labels map[string]string) *v1beta1.ReplicaSet {
		return &v1beta1.ReplicaSet{
			ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name},
			Spec: v1beta1.ReplicaSetSpec{
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: labels},
				},
			},
		}
	}

	testPodsFilter(t, genselector, "selector")
	testPodsFilter(t, gentemplate, "template")
}

func testPodsFilter(t *testing.T, gen func(string, string, map[string]string) *v1beta1.ReplicaSet, ctx string) {

	genpod := func(labels map[string]string) *v1.Pod {
		return &v1.Pod{ObjectMeta: metav1.ObjectMeta{Labels: labels}}
	}

	p1 := genpod(map[string]string{"a": "1", "b": "1", "c": "x"})
	p2 := genpod(map[string]string{"a": "2", "b": "2", "c": "x"})

	s1 := gen("a", "1", map[string]string{"a": "1"})
	s2 := gen("a", "2", map[string]string{"b": "2"})
	s3 := gen("c", "1", map[string]string{"c": "x"})
	s4 := gen("d", "1", map[string]string{"a": "0"})

	assert.False(t, replicaset.PodsFilter().Accept(p1), ctx)
	assert.True(t, replicaset.PodsFilter(s1).Accept(p1), ctx)
	assert.False(t, replicaset.PodsFilter(s1).Accept(p2), ctx)
	assert.True(t, replicaset.PodsFilter(s2).Accept(p2), ctx)
	assert.False(t, replicaset.PodsFilter(s2).Accept(p1), ctx)

	assert.True(t, replicaset.PodsFilter(s1, s2).Accept(p1), ctx)

	assert.True(t, replicaset.PodsFilter(s3).Accept(p1), ctx)
	assert.True(t, replicaset.PodsFilter(s3).Accept(p2), ctx)

	assert.False(t, replicaset.PodsFilter(s4).Accept(p1), ctx)
	assert.False(t, replicaset.PodsFilter(s4).Accept(p2), ctx)

	assert.True(t, replicaset.PodsFilter(s1).Equals(replicaset.PodsFilter(s1)), ctx)
	assert.True(t, replicaset.PodsFilter(s1, s2).Equals(replicaset.PodsFilter(s1, s2)), ctx)
	assert.True(t, replicaset.PodsFilter(s2, s1).Equals(replicaset.PodsFilter(s1, s2)), ctx)
	assert.True(t, replicaset.PodsFilter(s4, s3).Equals(replicaset.PodsFilter(s3, s4)), ctx)
}
