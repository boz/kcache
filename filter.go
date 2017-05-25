package kcache

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Filter interface {
	Accept(metav1.Object) bool
}

type nullFilter struct{}

func (nullFilter) Accept(_ metav1.Object) bool {
	return true
}

func NullFilter() Filter {
	return nullFilter{}
}
