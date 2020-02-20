package ingress

import (
	"github.com/boz/kcache/filter"
	"github.com/boz/kcache/nsname"
	appsv1 "k8s.io/api/networking/v1beta1"
)

func ServicesFilter(ingresses ...*appsv1.Ingress) filter.ComparableFilter {
	var ids []nsname.NSName

	for _, ing := range ingresses {
		ids = append(ids, buildServicesFilter(ing)...)
	}

	return filter.NSName(ids...)
}

func buildServicesFilter(ing *appsv1.Ingress) []nsname.NSName {
	var ids []nsname.NSName

	if be := ing.Spec.Backend; be != nil && be.ServiceName != "" {
		ids = append(ids, nsname.New(ing.GetNamespace(), be.ServiceName))
	}

	for _, rule := range ing.Spec.Rules {
		if http := rule.HTTP; http != nil {
			for _, path := range http.Paths {
				if service := path.Backend.ServiceName; service != "" {
					ids = append(ids, nsname.New(ing.GetNamespace(), service))
				}
			}
		}
	}

	return ids
}
