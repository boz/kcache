package kcache

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

func listResourceVersion(obj runtime.Object) (string, error) {
	list, err := meta.ListAccessor(obj)
	if err != nil {
		return "", err
	}
	return list.GetResourceVersion(), nil
}
