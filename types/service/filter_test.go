package service_test

import (
	"testing"

	"github.com/boz/kcache/types/service"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSelectorMatchFilter(t *testing.T) {
	target := map[string]string{"a": "1"}
	tsuper := map[string]string{"a": "1", "b": "2"}
	tmiss := map[string]string{"a": "2"}

	gensvc := func(labels map[string]string) metav1.Object {
		return &v1.Service{Spec: v1.ServiceSpec{Selector: labels}}
	}

	{
		f := service.SelectorMatchFilter(target)
		assert.True(t, f.Accept(gensvc(target)))
		assert.False(t, f.Accept(gensvc(tsuper)))
		assert.False(t, f.Accept(gensvc(tmiss)))
		assert.False(t, f.Accept(gensvc(nil)))

		assert.False(t, f.Accept(&v1.Pod{}))
	}

	{
		f := service.SelectorMatchFilter(tsuper)
		assert.True(t, f.Accept(gensvc(target)))
		assert.True(t, f.Accept(gensvc(tsuper)))
		assert.False(t, f.Accept(gensvc(tmiss)))
	}

	{
		f := service.SelectorMatchFilter(nil)
		assert.False(t, f.Accept(gensvc(target)))
	}

	{
		f := service.SelectorMatchFilter(target)
		fsuper := service.SelectorMatchFilter(tsuper)
		fnil := service.SelectorMatchFilter(nil)

		assert.True(t, f.Equals(f))
		assert.False(t, f.Equals(fsuper))
		assert.False(t, f.Equals(fnil))

		assert.True(t, fnil.Equals(fnil))
		assert.False(t, fnil.Equals(fsuper))
		assert.False(t, fnil.Equals(f))

		assert.True(t, fsuper.Equals(fsuper))
		assert.False(t, fsuper.Equals(fnil))
		assert.False(t, fsuper.Equals(f))
	}
}

func TestPodsFilter(t *testing.T) {

	genpod := func(labels map[string]string) *v1.Pod {
		return &v1.Pod{ObjectMeta: metav1.ObjectMeta{Labels: labels}}
	}

	gensvc := func(ns, name string, labels map[string]string) *v1.Service {
		return &v1.Service{
			ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name},
			Spec:       v1.ServiceSpec{Selector: labels},
		}
	}

	p1 := genpod(map[string]string{"a": "1", "b": "1", "c": "x"})
	p2 := genpod(map[string]string{"a": "2", "b": "2", "c": "x"})

	s1 := gensvc("a", "1", map[string]string{"a": "1"})
	s2 := gensvc("a", "2", map[string]string{"b": "2"})
	s3 := gensvc("c", "1", map[string]string{"c": "x"})
	s4 := gensvc("d", "1", map[string]string{"a": "0"})

	assert.False(t, service.PodsFilter().Accept(p1))
	assert.True(t, service.PodsFilter(s1).Accept(p1))
	assert.False(t, service.PodsFilter(s1).Accept(p2))
	assert.True(t, service.PodsFilter(s2).Accept(p2))
	assert.False(t, service.PodsFilter(s2).Accept(p1))

	assert.True(t, service.PodsFilter(s1, s2).Accept(p1))

	assert.True(t, service.PodsFilter(s3).Accept(p1))
	assert.True(t, service.PodsFilter(s3).Accept(p2))

	assert.False(t, service.PodsFilter(s4).Accept(p1))
	assert.False(t, service.PodsFilter(s4).Accept(p2))

	assert.True(t, service.PodsFilter(s1).Equals(service.PodsFilter(s1)))
	assert.True(t, service.PodsFilter(s1, s2).Equals(service.PodsFilter(s1, s2)))
	assert.True(t, service.PodsFilter(s2, s1).Equals(service.PodsFilter(s1, s2)))
	assert.True(t, service.PodsFilter(s4, s3).Equals(service.PodsFilter(s3, s4)))

}
