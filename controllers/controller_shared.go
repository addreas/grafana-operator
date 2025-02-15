package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var errorRequeueDelay = 60 * time.Second

func getMatchingInstances(ctx context.Context, k8sClient client.Client, labelSelector *v1.LabelSelector) (v1beta1.GrafanaList, error) {
	opts := []client.ListOption{}
	if labelSelector != nil {
		selector, err := v1.LabelSelectorAsSelector(labelSelector)
		if err != nil {
			return v1beta1.GrafanaList{}, err
		}
		opts = append(opts, client.MatchingLabelsSelector{Selector: selector})
	}
	var list v1beta1.GrafanaList

	if err := k8sClient.List(ctx, &list, opts...); err != nil {
		return v1beta1.GrafanaList{}, fmt.Errorf("failed to get grafana instances matching %v: %w", labelSelector, err)
	}

	if len(list.Items) == 0 {
		return v1beta1.GrafanaList{}, fmt.Errorf("no matching grafana instances matching: %v", labelSelector)
	}

	return list, nil
}

func updateGrafanaStatusPlugins(ctx context.Context, k8sClient client.Client, grafana *v1beta1.Grafana, plugins v1beta1.PluginList) error {
	log := log.FromContext(ctx, "grafana", client.ObjectKeyFromObject(grafana))
	plugins, err := grafana.Status.Plugins.ConsolidatedConcat(plugins)
	if err != nil {
		return fmt.Errorf("failed to extend plugin list with plugins %v: %w", plugins, err)
	}

	grafana.Status.Plugins = plugins
	log.Info("Updating plugins status for grafana", "plugins", plugins)
	if err := k8sClient.Status().Update(ctx, grafana); err != nil {
		return fmt.Errorf("failed to update plugin list in grafana status: %w", err)
	}

	return nil
}
