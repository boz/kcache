package filter_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/boz/kcache/filter"
	"github.com/boz/kcache/nsname"
)

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

func TestNotFilter(t *testing.T) {
	f1 := filter.Not(filter.All())
	f2 := filter.Not(filter.Null())

	assert.True(t, f1.Accept(&v1.Pod{}))
	assert.False(t, f2.Accept(&v1.Pod{}))

	assert.True(t, f1.Equals(f1))
	assert.False(t, f1.Equals(f2))
}

func TestNSName_fullset(t *testing.T) {
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

func TestNSName_partials(t *testing.T) {
	n1 := nsname.New("", "1")
	n2 := nsname.New("b", "")

	o1 := metav1.ObjectMeta{Namespace: "a", Name: "1"}
	o2 := metav1.ObjectMeta{Namespace: "b", Name: "2"}
	o3 := metav1.ObjectMeta{Namespace: "c", Name: "3"}

	assert.True(t, filter.NSName(n1).Accept(&v1.Pod{ObjectMeta: o1}))
	assert.False(t, filter.NSName(n1).Accept(&v1.Pod{ObjectMeta: o2}))
	assert.False(t, filter.NSName(n1).Accept(&v1.Pod{ObjectMeta: o3}))

	assert.True(t, filter.NSName(n1, n2).Accept(&v1.Service{ObjectMeta: o1}))
	assert.True(t, filter.NSName(n1, n2).Accept(&v1.Service{ObjectMeta: o2}))
	assert.False(t, filter.NSName(n1, n2).Accept(&v1.Service{ObjectMeta: o3}))

	assert.True(t, filter.NSName(n1).Equals(filter.NSName(n1)))
	assert.False(t, filter.NSName(n1).Equals(filter.NSName(n2)))
	assert.True(t, filter.NSName(n1, n2).Equals(filter.NSName(n1, n2)))
	assert.False(t, filter.NSName(n1, n2).Equals(filter.NSName(n2, n1)))
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

func TestFN(t *testing.T) {
	o1 := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "a", Name: "1"}}
	o2 := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Namespace: "b", Name: "1"}}

	f1 := filter.FN(func(obj metav1.Object) bool {
		return obj.GetNamespace() == "a"
	})

	assert.True(t, f1.Accept(o1))
	assert.False(t, f1.Accept(o2))
	assert.False(t, filter.FiltersEqual(f1, f1))
	assert.False(t, filter.FiltersEqual(f1, filter.All()))
}
