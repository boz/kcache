package event

import (
	"github.com/boz/kcache/client"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
)

const resourceName = "events"

func NewClient(cs kubernetes.Interface, ns string) client.Client {
	scope := cs.CoreV1()
	return client.ForResource(scope.RESTClient(), resourceName, ns, fields.Everything())
}
