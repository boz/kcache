package filter_test

import (
	"testing"

	"k8s.io/api/core/v1"

	"github.com/boz/kcache/filter"
	"github.com/stretchr/testify/assert"
)

func TestAndFilter(t *testing.T) {
	f := filter.And()
	assert.True(t, f.Accept(&v1.Pod{}))

	f = filter.And(filter.Null())
	assert.True(t, f.Accept(&v1.Pod{}))

	f = filter.And(filter.All())
	assert.False(t, f.Accept(&v1.Pod{}))

	f = filter.And(filter.Null(), filter.All())
	assert.False(t, f.Accept(&v1.Pod{}))

	a := filter.And()
	b := filter.And()
	assert.True(t, a.Equals(b))

	a = filter.And(filter.Null())
	b = filter.And()
	assert.False(t, a.Equals(b))

	a = filter.And()
	b = filter.And(filter.Null())
	assert.False(t, a.Equals(b))

	a = filter.And(filter.Null())
	b = filter.And(filter.Null())
	assert.True(t, a.Equals(b))

	a = filter.And(filter.Null())
	b = filter.And(filter.All())
	assert.False(t, a.Equals(b))
	assert.False(t, b.Equals(a))

	a = filter.And(filter.Null(), filter.All())
	b = filter.And(filter.All(), filter.Null())
	assert.False(t, a.Equals(b))
	assert.False(t, b.Equals(a))

	a = filter.And()
	b = filter.Or()
	assert.False(t, a.Equals(b))
}

func TestOrFilter(t *testing.T) {
	f := filter.Or()
	assert.False(t, f.Accept(&v1.Pod{}))

	f = filter.Or(filter.Null())
	assert.True(t, f.Accept(&v1.Pod{}))

	f = filter.Or(filter.All())
	assert.False(t, f.Accept(&v1.Pod{}))

	f = filter.Or(filter.Null(), filter.All())
	assert.True(t, f.Accept(&v1.Pod{}))

	a := filter.Or()
	b := filter.And()
	assert.False(t, a.Equals(b))
}
