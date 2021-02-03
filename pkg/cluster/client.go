package cluster

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetClient creates a new k8s client for use in talking to the cluster's
// api server.
func GetClient() (*kubernetes.Clientset, error) {
	log.Debug("initializing kubernetes client")

	config, err := GetKubeconfig()
	if err != nil {
		return nil, fmt.Errorf(
			"could not get cluster config: %s", err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("could not create client: %s", err)
	}

	log.WithFields(log.Fields{
		"host": config.Host,
	}).Debug("kubernetes client created")

	return client, nil
}

func GetKubeconfig() (*rest.Config, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	rules.DefaultClientConfig = &clientcmd.DefaultClientConfig

	overrides := &clientcmd.ConfigOverrides{}

	clientLoader := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		rules,
		overrides)

	return clientLoader.ClientConfig()
}
