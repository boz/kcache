package service_test

import (
	"testing"

	"github.com/boz/kcache/types/service"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
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
