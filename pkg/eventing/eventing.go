package eventing

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	eventingLabels = map[string]string{
		"app.kubernetes.io/instance": "eventing",
		"app.kubernetes.io/name":     "nats",
	}
)

func IsInstalled(config *rest.Config, logger logr.Logger) (bool, error) {
	k8sClient, err := client.New(config, client.Options{})
	if err != nil {
		return false, fmt.Errorf("failed to create Kubernetes Client: %v", err)
	}
	return isInstalledWithClient(k8sClient, logger)
}

func isInstalledWithClient(c client.Client, logger logr.Logger) (bool, error) {
	// use multiple label matches to be sure.
	matchingLabels := client.MatchingLabels(eventingLabels)
	listOpts := &client.ListOptions{}
	matchingLabels.ApplyToList(listOpts)

	stsList := &appsv1.StatefulSetList{}
	if err := c.List(context.Background(), stsList, listOpts); err != nil {
		return false, fmt.Errorf("failed to list statefulsets: %v", err)
	}

	if len(stsList.Items) > 0 {
		logger.Info(fmt.Sprintf("found [%d] statefulsets with matchingLabels: %v", len(stsList.Items), matchingLabels))
		return true, nil
	}

	return false, nil
}
