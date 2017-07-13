package secret

import (
	"github.com/boz/kcache"
	"k8s.io/client-go/pkg/api/v1"
)

func NewCache(parent kcache.CacheReader) CacheReader {
	return &cache{parent}
}

type cache struct {
	parent kcache.CacheReader
}

func (c *cache) Get(ns string, name string) (*v1.Secret, error) {
	obj, err := c.parent.Get(ns, name)
	if err != nil {
		return nil, err
	}
	return adapter.adaptObject(obj)
}

func (c *cache) List() ([]*v1.Secret, error) {
	objs, err := c.parent.List()
	if err != nil {
		return nil, err
	}
	return adapter.adaptList(objs)
}
