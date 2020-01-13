package job_test

import (
	"testing"

	"github.com/boz/kcache/types/job"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPodsFilter_selector(t *testing.T) {

	genselector := func(ns, name string, labels map[string]string) *batchv1.Job {
		return &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name},
			Spec: batchv1.JobSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: labels,
				},
			},
		}
	}

	gentemplate := func(ns, name string, labels map[string]string) *batchv1.Job {
		return &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name},
			Spec: batchv1.JobSpec{
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: labels},
				},
			},
		}
	}

	testPodsFilter(t, genselector, "selector")
	testPodsFilter(t, gentemplate, "template")
}

func testPodsFilter(t *testing.T, gen func(string, string, map[string]string) *batchv1.Job, ctx string) {

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

	assert.False(t, job.PodsFilter().Accept(p1), ctx)
	assert.True(t, job.PodsFilter(s1).Accept(p1), ctx)
	assert.False(t, job.PodsFilter(s1).Accept(p2), ctx)
	assert.True(t, job.PodsFilter(s2).Accept(p2), ctx)
	assert.False(t, job.PodsFilter(s2).Accept(p1), ctx)
	assert.False(t, job.PodsFilter(s5).Accept(p1), ctx)

	assert.True(t, job.PodsFilter(s1, s2).Accept(p1), ctx)

	assert.True(t, job.PodsFilter(s3).Accept(p1), ctx)
	assert.True(t, job.PodsFilter(s3).Accept(p2), ctx)

	assert.False(t, job.PodsFilter(s4).Accept(p1), ctx)
	assert.False(t, job.PodsFilter(s4).Accept(p2), ctx)

	assert.True(t, job.PodsFilter(s1).Equals(job.PodsFilter(s1)), ctx)
	assert.True(t, job.PodsFilter(s1, s2).Equals(job.PodsFilter(s1, s2)), ctx)
	assert.True(t, job.PodsFilter(s2, s1).Equals(job.PodsFilter(s1, s2)), ctx)
	assert.True(t, job.PodsFilter(s4, s3).Equals(job.PodsFilter(s3, s4)), ctx)

}
