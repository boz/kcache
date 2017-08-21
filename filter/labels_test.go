package filter_test

import (
	"testing"

	"github.com/boz/kcache/filter"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLabels(t *testing.T) {

	target := map[string]string{"a": "1"}
	tsuper := map[string]string{"a": "1", "b": "2"}
	tmiss := map[string]string{"a": "2"}

	f := filter.Labels(target)
	fnil := filter.Labels(nil)

	gen := func(labels map[string]string) metav1.Object {
		return &v1.Pod{ObjectMeta: metav1.ObjectMeta{Labels: labels}}
	}

	assert.True(t, f.Accept(gen(target)))
	assert.True(t, f.Accept(gen(tsuper)))
	assert.False(t, f.Accept(gen(tmiss)))

	assert.True(t, fnil.Accept(gen(target)))
	assert.True(t, fnil.Accept(gen(tsuper)))
	assert.True(t, fnil.Accept(gen(tmiss)))

	assert.True(t, f.Equals(f))
	assert.False(t, f.Equals(fnil))

	assert.True(t, fnil.Equals(fnil))
	assert.False(t, fnil.Equals(f))

	fempty := filter.Labels(map[string]string{})
	assert.True(t, fempty.Accept(gen(target)))
	assert.True(t, fempty.Accept(gen(tsuper)))
	assert.True(t, fempty.Accept(gen(tmiss)))
	assert.True(t, fempty.Accept(gen(map[string]string{})))
	assert.True(t, fempty.Equals(fempty))
}

func TestLabelSelector(t *testing.T) {
	target := map[string]string{"a": "1"}
	tsuper := map[string]string{"a": "1", "b": "2"}
	tmiss := map[string]string{"a": "2"}

	gen := func(labels map[string]string) metav1.Object {
		return &v1.Pod{ObjectMeta: metav1.ObjectMeta{Labels: labels}}
	}

	fnil := filter.LabelSelector(nil)
	fempty := filter.LabelSelector(&metav1.LabelSelector{})
	fmatch := filter.LabelSelector(&metav1.LabelSelector{MatchLabels: target})

	fexpr := filter.LabelSelector(&metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{Key: "a", Operator: metav1.LabelSelectorOpIn, Values: []string{"1"}},
		},
	})

	assert.False(t, fnil.Accept(gen(target)))
	assert.True(t, fempty.Accept(gen(target)))
	assert.True(t, fmatch.Accept(gen(target)))
	assert.True(t, fexpr.Accept(gen(target)))

	assert.False(t, fnil.Accept(gen(tsuper)))
	assert.True(t, fempty.Accept(gen(tsuper)))
	assert.True(t, fmatch.Accept(gen(tsuper)))
	assert.True(t, fexpr.Accept(gen(tsuper)))

	assert.False(t, fnil.Accept(gen(tmiss)))
	assert.True(t, fempty.Accept(gen(tmiss)))
	assert.False(t, fmatch.Accept(gen(tmiss)))
	assert.False(t, fexpr.Accept(gen(tmiss)))

	assert.True(t, fnil.Equals(fnil))
	assert.True(t, fempty.Equals(fempty))
	assert.True(t, fmatch.Equals(fmatch))
	assert.True(t, fexpr.Equals(fexpr))

	assert.False(t, fnil.Equals(fempty))
	assert.False(t, fempty.Equals(fnil))

	assert.False(t, fnil.Equals(fmatch))
	assert.False(t, fnil.Equals(fmatch))
	assert.False(t, fmatch.Equals(fexpr))
	assert.False(t, fexpr.Equals(fmatch))
	assert.False(t, fexpr.Equals(filter.All()))
}
