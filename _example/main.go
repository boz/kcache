package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	logutil "github.com/boz/go-logutil"
	"github.com/boz/kcache"
	"github.com/boz/kcache/types/pod"

	lr "github.com/boz/go-logutil/logrus"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.Level = logrus.DebugLevel

	log := lr.New(logger)
	ctx := context.Background()

	cs := getClientset(log)

	client := pod.NewClient(cs, metav1.NamespaceAll)

	controller, err := kcache.NewController(ctx, log, client)
	if err != nil {
		log.ErrFatal(err, "kcache.NewController()")
	}

	go watchSignals(log, controller)

	defer controller.Close()

	subscription, err := controller.Subscribe()
	if err != nil {
		log.ErrFatal(err, "subscribe")
	}

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
		fmt.Printf("%v/%v: %v\n", pod.GetNamespace(), pod.GetName(), pod.GetResourceVersion())
	}

	for {
		select {
		case ev, ok := <-subscription.Events():
			if !ok {
				return
			}
			obj := ev.Resource()
			fmt.Printf("event: %v: %v/%v[%v]\n", ev.Type(), obj.GetNamespace(), obj.GetName(), obj.GetResourceVersion())

			cnobj, err := subscription.Cache().Get(obj.GetNamespace(), obj.GetName())
			if err != nil {
				log.ErrWarn(err, "Get()")
				continue
			}

			if ev.Type() == kcache.EventTypeDelete {
				if cnobj != nil {
					log.Warnf("Get(deleted) != nil")
				}
				continue
			}

			if cnobj == nil {
				log.Warnf("Get() -> nil")
				continue
			}

			fmt.Printf("Get: %v/%v[%v]\n", cnobj.GetNamespace(), cnobj.GetName(), cnobj.GetResourceVersion())
		}
	}
}

func getRESTClient(log logutil.Log) rest.Interface {
	clientset := getClientset(log)

	client := clientset.Core().RESTClient()
	return client
}

func getClientset(log logutil.Log) *kubernetes.Clientset {
	kconfig, err := getKubeRESTConfig()
	if err != nil {
		log.ErrFatal(err, "can't get kube client")
	}

	clientset, err := kubernetes.NewForConfig(kconfig)
	if err != nil {
		log.ErrFatal(err, "can't get clientset")
	}
	return clientset
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
