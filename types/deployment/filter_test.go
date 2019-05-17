package deployment_test

import (
	"testing"

	"github.com/boz/kcache/types/deployment"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPodsFilter_selector(t *testing.T) {

	genselector := func(ns, name string, labels map[string]string) *v1beta1.Deployment {
		return &v1beta1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name},
			Spec: v1beta1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: labels,
				},
			},
		}
	}

	gentemplate := func(ns, name string, labels map[string]string) *v1beta1.Deployment {
		return &v1beta1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name},
			Spec: v1beta1.DeploymentSpec{
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: labels},
				},
			},
		}
	}

	testPodsFilter(t, genselector, "selector")
	testPodsFilter(t, gentemplate, "template")
}

func testPodsFilter(t *testing.T, gen func(string, string, map[string]string) *v1beta1.Deployment, ctx string) {

	genpod := func(ns string, labels map[string]string) *v1.Pod {
		return &v1.Pod{ObjectMeta: metav1.ObjectMeta{Labels: labels, Namespace: ns}}
	}

	p1 := genpod("a", map[string]string{"a": "1", "b": "1", "c": "x"})
	p2 := genpod("a", map[string]string{"a": "2", "b": "2", "c": "x"})

	s1 := gen("a", "1", map[string]string{"a": "1"})
	s2 := gen("a", "2", map[string]string{"b": "2"})
	s3 := gen("a", "3", map[string]string{"c": "x"})
	s4 := gen("a", "4", map[string]string{"a": "0"})
	s5 := gen("b", "1", map[string]string{"a": "1"})

	assert.False(t, deployment.PodsFilter().Accept(p1), ctx)
	assert.True(t, deployment.PodsFilter(s1).Accept(p1), ctx)
	assert.False(t, deployment.PodsFilter(s1).Accept(p2), ctx)
	assert.True(t, deployment.PodsFilter(s2).Accept(p2), ctx)
	assert.False(t, deployment.PodsFilter(s2).Accept(p1), ctx)
	assert.False(t, deployment.PodsFilter(s5).Accept(p1), ctx)

	assert.True(t, deployment.PodsFilter(s1, s2).Accept(p1), ctx)

	assert.True(t, deployment.PodsFilter(s3).Accept(p1), ctx)
	assert.True(t, deployment.PodsFilter(s3).Accept(p2), ctx)

	assert.False(t, deployment.PodsFilter(s4).Accept(p1), ctx)
	assert.False(t, deployment.PodsFilter(s4).Accept(p2), ctx)

	assert.True(t, deployment.PodsFilter(s1).Equals(deployment.PodsFilter(s1)), ctx)
	assert.True(t, deployment.PodsFilter(s1, s2).Equals(deployment.PodsFilter(s1, s2)), ctx)
	assert.True(t, deployment.PodsFilter(s2, s1).Equals(deployment.PodsFilter(s1, s2)), ctx)
	assert.True(t, deployment.PodsFilter(s4, s3).Equals(deployment.PodsFilter(s3, s4)), ctx)

}
