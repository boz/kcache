package replicaset

import (
	"sort"

	"k8s.io/api/extensions/v1beta1"

	"github.com/boz/kcache/filter"
)

func PodsFilter(sources ...*v1beta1.ReplicaSet) filter.ComparableFilter {

	// make a copy and sort
	srcs := make([]*v1beta1.ReplicaSet, len(sources))
	copy(srcs, sources)

	sort.Slice(srcs, func(i, j int) bool {
		if srcs[i].Namespace != srcs[j].Namespace {
			return srcs[i].Namespace < srcs[j].Namespace
		}
		return srcs[i].Name < srcs[j].Name
	})

	filters := make([]filter.Filter, 0, len(srcs))

	for _, svc := range srcs {

		// TODO: match expressions
		var labels map[string]string
		if svc.Spec.Selector != nil {
			labels = svc.Spec.Selector.MatchLabels
		} else {
			labels = svc.Spec.Template.Labels
		}

		filters = append(filters, filter.Labels(labels))
	}

	return filter.Or(filters...)
}
