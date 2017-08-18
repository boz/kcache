package ingress_test

import (
	"testing"

	"github.com/boz/kcache/types/ingress"
	"github.com/stretchr/testify/assert"

	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestServicesFilter(t *testing.T) {

	gensvc := func(ns, name string) *v1.Service {
		return &v1.Service{
			ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name},
		}
	}

	ing1 := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{Namespace: "a", Name: "1"},
		Spec: v1beta1.IngressSpec{
			Backend: &v1beta1.IngressBackend{
				ServiceName: "foo",
			},
		},
	}

	ing2 := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{Namespace: "b", Name: "2"},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					IngressRuleValue: v1beta1.IngressRuleValue{
						HTTP: &v1beta1.HTTPIngressRuleValue{
							Paths: []v1beta1.HTTPIngressPath{
								{
									Backend: v1beta1.IngressBackend{
										ServiceName: "bar",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	assert.True(t, ingress.ServicesFilter(ing1).Accept(gensvc("a", "foo")))
	assert.False(t, ingress.ServicesFilter(ing1).Accept(gensvc("a", "bar")))
	assert.False(t, ingress.ServicesFilter(ing1).Accept(gensvc("b", "foo")))

	assert.True(t, ingress.ServicesFilter(ing2).Accept(gensvc("b", "bar")))
	assert.False(t, ingress.ServicesFilter(ing2).Accept(gensvc("b", "foo")))
	assert.False(t, ingress.ServicesFilter(ing2).Accept(gensvc("a", "bar")))

	assert.True(t, ingress.ServicesFilter(ing1, ing2).Accept(gensvc("a", "foo")))
	assert.True(t, ingress.ServicesFilter(ing1, ing2).Accept(gensvc("b", "bar")))

	assert.True(t, ingress.ServicesFilter(ing1).Equals(ingress.ServicesFilter(ing1)))
	assert.False(t, ingress.ServicesFilter(ing1).Equals(ingress.ServicesFilter(ing2)))
	assert.True(t, ingress.ServicesFilter(ing1, ing2).Equals(ingress.ServicesFilter(ing2, ing1)))

}
