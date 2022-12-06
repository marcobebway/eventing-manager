/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/eventing-manager/api/v1alpha1"
	"github.com/kyma-project/module-manager/pkg/declarative"
	"github.com/kyma-project/module-manager/pkg/types"
)

const (
	chartNamespace  = "kyma-system"
	pruneLabelKey   = "eventing-manager.kyma-project.io/prune"
	pruneLabelValue = "true"
	finalizer       = "eventing-manager.kyma-project.io/deletion-hook"
)

// EventingReconciler reconciles a Eventing object
type EventingReconciler struct {
	declarative.ManifestReconciler
	Scheme    *runtime.Scheme
	ChartPath string
	client.Client
	*rest.Config
}

//+kubebuilder:rbac:groups="*",resources="*",verbs=get
//+kubebuilder:rbac:groups=external.metrics.k8s.io,resources="*",verbs="*"
//+kubebuilder:rbac:groups="",resources=configmaps;configmaps/status;events;services,verbs="*"
//+kubebuilder:rbac:groups="",resources=external;pods;secrets;serviceaccounts,verbs=list;watch;create;delete;update;patch
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=create;delete
//+kubebuilder:rbac:groups=apiregistration.k8s.io,resources=apiservices,verbs=create;delete;update;patch
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterrolebindings;clusterroles;rolebindings,verbs=create;delete;update;patch
//+kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=create;delete;update;patch
//+kubebuilder:rbac:groups="*",resources="*/scale",verbs="*"
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=list;watch;create;delete;update;patch
//+kubebuilder:rbac:groups=apps,resources=statefulsets;replicasets,verbs=list;watch;create;delete;update;patch
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs="*"
//+kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs="*"
//+kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs="*"
//+kubebuilder:rbac:groups=operator.kyma-project.io,resources=eventings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=operator.kyma-project.io,resources=eventings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=operator.kyma-project.io,resources=eventings/finalizers,verbs=update;patch

// initReconciler injects the required configuration into the declarative reconciler.
func (r *EventingReconciler) initReconciler(mgr ctrl.Manager) error {
	manifestResolver := &ManifestResolver{chartPath: r.ChartPath}
	return r.Inject(mgr, &v1alpha1.Eventing{},
		declarative.WithResourcesReady(true),
		declarative.WithFinalizer(finalizer),
		declarative.WithManifestResolver(manifestResolver),
		withPruneStatefulSetsByLabels(map[string]string{pruneLabelKey: pruneLabelValue}),
	)
}

// SetupWithManager sets up the controller with the Manager.
func (r *EventingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Config = mgr.GetConfig()
	if err := r.initReconciler(mgr); err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).For(&v1alpha1.Eventing{}).Complete(r)
}

func structToFlags(obj interface{}) (types.Flags, error) {
	data, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	var flags types.Flags
	if err = json.Unmarshal(data, &flags); err != nil {
		return nil, err
	}

	return flags, nil
}

// ManifestResolver represents the chart information for the passed Sample resource.
type ManifestResolver struct {
	chartPath string
}

// Get returns the chart information to be processed.
func (m *ManifestResolver) Get(obj types.BaseCustomObject, l logr.Logger) (types.InstallationSpec, error) {
	eventing, valid := obj.(*v1alpha1.Eventing)
	if !valid {
		return types.InstallationSpec{}, fmt.Errorf("invalid type conversion for %s", client.ObjectKeyFromObject(obj))
	}

	flags, err := structToFlags(eventing.Spec)
	if err != nil {
		return types.InstallationSpec{}, fmt.Errorf("resolving manifest failed: %w", err)
	}

	l.Info("Eventing",
		"flags", flags,
		"backend", eventing.Spec.BackendSpec.Type,
	)

	return types.InstallationSpec{
		ChartPath: m.chartPath,
		ChartFlags: types.ChartFlags{
			ConfigFlags: types.Flags{
				"Namespace":       chartNamespace,
				"CreateNamespace": true,
			},
			SetFlags: types.Flags{
				"nats": map[string]interface{}{
					"enabled": eventing.Spec.BackendSpec.Type == v1alpha1.BackendTypeNats,
				},
			},
		},
	}, nil
}

func withPruneStatefulSetsByLabels(labelSet labels.Set) declarative.ReconcilerOption {
	return declarative.With(
		declarative.WithPostRenderTransform(labelInjector(labelSet)),
		declarative.WithPostRun(pruneStatefulSetsByLabels(labelSet)),
	)
}

func labelInjector(labelSet labels.Set) types.ObjectTransform {
	return func(_ context.Context, _ types.BaseCustomObject, resources *types.ManifestResources) error {
		for i := range resources.Items {
			if resources.Items[i].GroupVersionKind().Kind == "StatefulSet" {
				ls := resources.Items[i].GetLabels()
				if ls == nil {
					ls = map[string]string{}
				}
				for key := range labelSet {
					ls[key] = labelSet[key]
				}
				resources.Items[i].SetLabels(ls)
			}
		}
		return nil
	}
}

func pruneStatefulSetsByLabels(labelSet labels.Set) types.PostRun {
	return func(ctx context.Context, c client.Client, obj types.BaseCustomObject, _ types.ResourceLists) error {
		eventing := &v1alpha1.Eventing{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.(*unstructured.Unstructured).Object, eventing); err != nil {
			return fmt.Errorf("invalid type conversion for %s", client.ObjectKeyFromObject(obj))
		}

		if eventing.Spec.BackendSpec.Type == v1alpha1.BackendTypeEventMesh {
			statefulSetList := &v1.StatefulSetList{}
			listOptions := &client.ListOptions{LabelSelector: labels.SelectorFromSet(labelSet)}
			if err := c.List(ctx, statefulSetList, listOptions); err != nil {
				return fmt.Errorf("failed to list statefulsets to prune when switching to eventmesh: %w", err)
			}

			propagationPolicy := metav1.DeletePropagationForeground
			deleteOptions := &client.DeleteOptions{PropagationPolicy: &propagationPolicy}
			for i := range statefulSetList.Items {
				if err := c.Delete(ctx, &statefulSetList.Items[i], deleteOptions); err != nil {
					return fmt.Errorf("failed to prune statefulsets %s when switching to eventmesh: %w",
						client.ObjectKeyFromObject(&statefulSetList.Items[i]), err)
				}
			}
		}

		return nil
	}
}
