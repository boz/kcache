package filter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/boz/kcache/filter"
	"github.com/boz/kcache/nsname"
)

type filterTests []struct {
	object metav1.Object
	result bool
}

func TestNullFilter(t *testing.T) {
	f := filter.Null()

	assert.True(t, f.Accept(&v1.Pod{}))
	assert.True(t, f.Accept(&v1.Service{}))
	assert.True(t, f.Accept(&v1.Secret{}))

	assert.True(t, f.Equals(filter.Null()))
	assert.False(t, f.Equals(nil))
	assert.False(t, f.Equals(filter.All()))
}

func TestAllFilter(t *testing.T) {
	f := filter.All()

	assert.False(t, f.Accept(&v1.Pod{}))
	assert.False(t, f.Accept(&v1.Service{}))
	assert.False(t, f.Accept(&v1.Secret{}))

	assert.True(t, f.Equals(filter.All()))
	assert.False(t, f.Equals(nil))
	assert.False(t, f.Equals(filter.Null()))
}

func TestLabels(t *testing.T) {
	target := map[string]string{"a": "1"}
	tsuper := map[string]string{"a": "1", "b": "2"}
	tmiss := map[string]string{"a": "2"}

	f := filter.Labels(target)
	fempty := filter.Labels(nil)

	gen := func(labels map[string]string) metav1.Object {
		return &v1.Pod{ObjectMeta: metav1.ObjectMeta{Labels: labels}}
	}

	assert.True(t, f.Accept(gen(target)))
	assert.True(t, f.Accept(gen(tsuper)))
	assert.False(t, f.Accept(gen(tmiss)))

	assert.True(t, fempty.Accept(gen(target)))
	assert.True(t, fempty.Accept(gen(tsuper)))
	assert.True(t, fempty.Accept(gen(tmiss)))

	assert.True(t, f.Equals(f))
	assert.False(t, f.Equals(fempty))

	assert.True(t, fempty.Equals(fempty))
	assert.False(t, fempty.Equals(f))
}

func TestServiceFor(t *testing.T) {
	target := map[string]string{"a": "1"}
	tsuper := map[string]string{"a": "1", "b": "2"}
	tmiss := map[string]string{"a": "2"}

	gensvc := func(labels map[string]string) metav1.Object {
		return &v1.Service{Spec: v1.ServiceSpec{Selector: labels}}
	}

	{
		f := filter.ServiceFor(target)
		assert.True(t, f.Accept(gensvc(target)))
		assert.False(t, f.Accept(gensvc(tsuper)))
		assert.False(t, f.Accept(gensvc(tmiss)))
		assert.False(t, f.Accept(gensvc(nil)))

		assert.False(t, f.Accept(&v1.Pod{}))
	}

	{
		f := filter.ServiceFor(tsuper)
		assert.True(t, f.Accept(gensvc(target)))
		assert.True(t, f.Accept(gensvc(tsuper)))
		assert.False(t, f.Accept(gensvc(tmiss)))
	}

	{
		f := filter.ServiceFor(nil)
		assert.False(t, f.Accept(gensvc(target)))
	}

	{
		f := filter.ServiceFor(target)
		fsuper := filter.ServiceFor(tsuper)
		fnil := filter.ServiceFor(nil)

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

func TestNSName(t *testing.T) {
	n1 := nsname.New("a", "1")
	n2 := nsname.New("a", "2")
	n3 := nsname.New("b", "2")

	o1 := metav1.ObjectMeta{Namespace: n1.Namespace, Name: n1.Name}
	o2 := metav1.ObjectMeta{Namespace: n2.Namespace, Name: n2.Name}
	o3 := metav1.ObjectMeta{Namespace: n3.Namespace, Name: n3.Name}

	assert.True(t, filter.NSName(n1).Accept(&v1.Pod{ObjectMeta: o1}))
	assert.False(t, filter.NSName(n1).Accept(&v1.Pod{ObjectMeta: o2}))
	assert.False(t, filter.NSName(n1).Accept(&v1.Pod{ObjectMeta: o3}))

	assert.True(t, filter.NSName(n1, n2).Accept(&v1.Service{ObjectMeta: o1}))
	assert.True(t, filter.NSName(n1, n2).Accept(&v1.Service{ObjectMeta: o2}))
	assert.False(t, filter.NSName(n1, n2).Accept(&v1.Service{ObjectMeta: o3}))

	assert.False(t, filter.NSName().Accept(&v1.Service{ObjectMeta: o1}))

	assert.True(t, filter.NSName().Equals(filter.NSName()))
	assert.True(t, filter.NSName(n1).Equals(filter.NSName(n1)))
	assert.False(t, filter.NSName(n1).Equals(filter.NSName(n2)))
	assert.False(t, filter.NSName(n1).Equals(nil))
	assert.False(t, filter.NSName().Equals(nil))
}

func TestFiltersEqual(t *testing.T) {

	assert.True(t, filter.FiltersEqual(nil, nil))
	assert.True(t, filter.FiltersEqual(filter.Null(), filter.Null()))
	assert.False(t, filter.FiltersEqual(filter.Null(), nil))
	assert.False(t, filter.FiltersEqual(filter.Null(), filter.All()))
	assert.False(t, filter.FiltersEqual(filter.All(), filter.Null()))

	assert.True(t, filter.FiltersEqual(filter.NSName(nsname.New("a", "1")), filter.NSName(nsname.New("a", "1"))))
	assert.False(t, filter.FiltersEqual(filter.NSName(nsname.New("a", "1")), filter.NSName(nsname.New("a", "2"))))
}
