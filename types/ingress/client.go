package ingress

import (
	"github.com/boz/kcache/client"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
)

func NewClient(cs kubernetes.Interface, ns string) client.Client {
	return client.ForResource(
		cs.ExtensionsV1beta1().RESTClient(), "ingresses", ns, fields.Everything())
}
