package pod_test

import (
	"testing"

	"github.com/boz/kcache/types/pod"
	"github.com/stretchr/testify/assert"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNodeFilter(t *testing.T) {

	genpod := func(name, node string) *v1.Pod {
		return &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: name},
			Spec:       v1.PodSpec{NodeName: node},
		}
	}

	assert.True(t, pod.NodeFilter("a").Accept(genpod("x", "a")))
	assert.True(t, pod.NodeFilter("a", "b").Accept(genpod("x", "a")))
	assert.False(t, pod.NodeFilter().Accept(genpod("x", "a")))
	assert.False(t, pod.NodeFilter("a").Accept(genpod("x", "b")))
	assert.False(t, pod.NodeFilter("a", "c").Accept(genpod("x", "b")))

	assert.True(t, pod.NodeFilter().Equals(pod.NodeFilter()))
	assert.True(t, pod.NodeFilter("a").Equals(pod.NodeFilter("a")))
	assert.True(t, pod.NodeFilter("a", "b").Equals(pod.NodeFilter("a", "b")))
	assert.True(t, pod.NodeFilter("b", "a").Equals(pod.NodeFilter("a", "b")))

	other := otherFilter(make(map[string]interface{}))
	assert.False(t, pod.NodeFilter().Equals(other))
}

type otherFilter map[string]interface{}

func (otherFilter) Accept(_ metav1.Object) bool {
	return false
}
