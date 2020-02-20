package replicaset

import (
	"sort"

	appsv1 "k8s.io/api/apps/v1"

	"github.com/boz/kcache/filter"
	"github.com/boz/kcache/nsname"
)

func PodsFilter(sources ...*appsv1.ReplicaSet) filter.ComparableFilter {

	// make a copy and sort
	srcs := make([]*appsv1.ReplicaSet, len(sources))
	copy(srcs, sources)

	sort.Slice(srcs, func(i, j int) bool {
		if srcs[i].Namespace != srcs[j].Namespace {
			return srcs[i].Namespace < srcs[j].Namespace
		}
		return srcs[i].Name < srcs[j].Name
	})

	filters := make([]filter.Filter, 0, len(srcs))

	for _, svc := range srcs {

		var sfilter filter.Filter
		if sel := svc.Spec.Selector; sel != nil {
			sfilter = filter.LabelSelector(sel)
		} else {
			sfilter = filter.Labels(svc.Spec.Template.Labels)
		}

		nsfilter := filter.NSName(nsname.New(svc.GetNamespace(), ""))

		filters = append(filters, filter.And(nsfilter, sfilter))
	}

	return filter.Or(filters...)
}
