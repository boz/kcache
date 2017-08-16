package util

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func KubeClient(overrides *clientcmd.ConfigOverrides) (kubernetes.Interface, *rest.Config, error) {
	config, err := KubeConfig(overrides)
	if err != nil {
		return nil, nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	return clientset, config, nil
}

func KubeConfig(overrides *clientcmd.ConfigOverrides) (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, err
	}

	if overrides == nil {
		overrides = &clientcmd.ConfigOverrides{}
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		overrides,
	).ClientConfig()
}
