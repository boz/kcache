/*
 * AUTO GENERATED - DO NOT EDIT BY HAND
 */

package join

import (
	"context"

	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache/filter"
	"github.com/boz/kcache/types/deployment"
	"github.com/boz/kcache/types/pod"
	appsv1 "k8s.io/api/apps/v1"
)

func DeploymentPodsWith(ctx context.Context,
	srcController deployment.Controller,
	dstController pod.Publisher,
	filterFn func(...*appsv1.Deployment) filter.ComparableFilter) (pod.Controller, error) {

	log := logutil.FromContextOrDefault(ctx)

	dst, err := dstController.CloneForFilter()
	if err != nil {
		return nil, err
	}

	update := func(_ *appsv1.Deployment) {
		objs, err := srcController.Cache().List()
		if err != nil {
			log.Err(err, "join(deployment,pod: cache list")
			return
		}
		dst.Refilter(filterFn(objs...))
	}

	handler := deployment.BuildHandler().
		OnInitialize(func(objs []*appsv1.Deployment) { dst.Refilter(filterFn(objs...)) }).
		OnCreate(update).
		OnUpdate(update).
		OnDelete(update).
		Create()

	monitor, err := deployment.NewMonitor(srcController, handler)
	if err != nil {
		dst.Close()
		return nil, log.Err(err, "join(deployment,pod): monitor")
	}

	go func() {
		<-dst.Done()
		monitor.Close()
	}()

	return dst, nil
}
