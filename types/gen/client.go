package main

import (
	"github.com/boz/kcache/client"
	"k8s.io/client-go/kubernetes"
)

func NewClient(cs kubernetes.Interface, ns string) client.Client {
	return client.ForResource(
		cs.CoreV1().RESTClient(), "pods", ns)
}
