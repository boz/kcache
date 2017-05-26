package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache"
	"github.com/boz/kcache/client"

	lr "github.com/boz/go-logutil/logrus"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel

	log := lr.New(logger)
	ctx := context.Background()

	rclient := getRESTClient(log)

	client := client.ForResource(rclient, "pods", metav1.NamespaceAll, fields.Everything())

	controller, err := kcache.NewController(ctx, log, client)
	if err != nil {
		log.ErrFatal(err, "kcache.NewController()")
	}

	go watchSignals(log, controller)

	defer controller.Close()

	subscription := controller.Subscribe()

	select {
	case <-subscription.Ready():
	case <-subscription.Done():
		return
	}

	list, err := subscription.Cache().List()
	if err != nil {
		log.ErrFatal(err, "Cache().List()")
	}

	for _, pod := range list {
		if pod, ok := pod.(*v1.Pod); ok {
			fmt.Printf("%v/%v: %v\n", pod.GetNamespace(), pod.GetName(), pod.GetResourceVersion())
		} else {
			log.Infof("invalid type: %T", pod)
		}
	}

	for {
		select {
		case ev, ok := <-subscription.Events():
			if !ok {
				return
			}
			obj := ev.Resource()
			fmt.Printf("event: %v: %v/%v[%v]\n", ev.Type(), obj.GetNamespace(), obj.GetName(), obj.GetResourceVersion())

			cobj, err := subscription.Cache().GetObject(obj)
			if err != nil {
				log.ErrWarn(err, "GetObject()")
				continue
			}

			cnobj, err := subscription.Cache().Get(obj.GetNamespace(), obj.GetName())
			if err != nil {
				log.ErrWarn(err, "Get()")
				continue
			}

			if ev.Type() == kcache.EventTypeDelete {
				if cobj != nil {
					log.Warnf("GetObject(deleted) != nil")
				}
				if cnobj != nil {
					log.Warnf("Get(deleted) != nil")
				}
				continue
			}

			if cobj == nil {
				log.Warnf("GetObject() -> nil")
				continue
			}
			if cnobj == nil {
				log.Warnf("Get() -> nil")
				continue
			}

			fmt.Printf("GetObject: %v/%v[%v]\n", cobj.GetNamespace(), cobj.GetName(), cobj.GetResourceVersion())
			fmt.Printf("Get: %v/%v[%v]\n", cnobj.GetNamespace(), cnobj.GetName(), cnobj.GetResourceVersion())
		}
	}
}

func getRESTClient(log logutil.Log) rest.Interface {
	kconfig, err := getKubeRESTConfig()
	if err != nil {
		log.ErrFatal(err, "can't get kube client")
	}

	clientset, err := kubernetes.NewForConfig(kconfig)
	if err != nil {
		log.ErrFatal(err, "can't get clientset")
	}

	client := clientset.Core().RESTClient()
	return client
}

func getKubeRESTConfig() (*rest.Config, error) {
	/*
		config, err := rest.InClusterConfig()
		if err == nil {
			return config, err
		}
	*/

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
}

func watchSignals(log logutil.Log, controller kcache.Controller) {
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGQUIT)

	select {
	case <-controller.Done():
	case <-sigch:
		controller.Close()
	}
}
