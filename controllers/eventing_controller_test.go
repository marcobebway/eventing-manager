package controllers

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kyma-project/eventing-manager/api/v1alpha1"
	rtypes "github.com/kyma-project/module-manager/pkg/types"
)

var _ = Describe("Eventing controller", func() {
	Context("When creating fresh instance", func() {
		const (
			namespaceName   = "kyma-system"
			eventingName    = "test"
			statefulSetName = "eventing-nats"
		)

		var (
			EventingSpec = v1alpha1.EventingSpec{
				BackendSpec: v1alpha1.BackendSpec{
					Type: v1alpha1.BackendTypeEventMesh,
				},
			}
		)

		It("The status should be Success", func() {
			h := testHelper{
				ctx:           context.Background(),
				namespaceName: namespaceName,
			}
			h.createNamespace()

			// operations like C(R)UD can be tested in separated tests,
			// but we have time-consuming flow and decided do it in one test
			shouldCreateEventing(h, eventingName, EventingSpec)

			shouldUpdateEventing(h, eventingName, statefulSetName)

			shouldPropagateEventingCrdSpecProperties(h, statefulSetName)

			shouldDeleteEventing(h, eventingName)
		})
	})
})

func shouldCreateEventing(h testHelper, eventingName string, eventingSpec v1alpha1.EventingSpec) {
	// act
	h.createEventing(eventingName, eventingSpec)

	// assert
	Eventually(h.createGetEventingStateFunc(eventingName)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 20).
		Should(Equal(rtypes.StateReady))
}

func shouldUpdateEventing(h testHelper, eventingName, eventingStatefulSetName string) {
	// arrange
	var eventing v1alpha1.Eventing
	Eventually(h.createGetKubernetesObjectFunc(eventingName, &eventing)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 10).
		Should(BeTrue())

	eventing.Spec.BackendSpec.Type = v1alpha1.BackendTypeNats

	// act
	Expect(k8sClient.Update(h.ctx, &eventing)).To(Succeed())

	// assert
	Eventually(h.createGetKubernetesObjectFunc(eventingName, &eventing)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 10).
		Should(BeTrue())

	Eventually(h.createGetEventingStateFunc(eventingName)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 20).
		Should(Equal(rtypes.StateReady))

	// we have to update statefulSet status manually
	h.updateStatefulSetStatus(eventingStatefulSetName)
}

func shouldDeleteEventing(h testHelper, eventingName string) {
	// initial assert
	Expect(h.getEventingCount()).To(Equal(1))
	Eventually(h.createGetEventingStateFunc(eventingName)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 20).
		Should(Equal(rtypes.StateReady))

	// act
	var eventing v1alpha1.Eventing
	Eventually(h.createGetKubernetesObjectFunc(eventingName, &eventing)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 10).
		Should(BeTrue())
	Expect(k8sClient.Delete(h.ctx, &eventing)).To(Succeed())

	// assert
	Eventually(h.getEventingCount).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 10).
		Should(Equal(1)) // Change to zero after
	// issue: https://github.com/kyma-project/module-manager/issues/191 is resolved
}

func shouldPropagateEventingCrdSpecProperties(h testHelper, eventingStatefulSetName string) {
	checkEventingCrdSpecPropertyPropagationToEventingStatefulSet(h, eventingStatefulSetName)
}

func checkEventingCrdSpecPropertyPropagationToEventingStatefulSet(h testHelper, eventingStatefulSetName string) {
	// act
	var eventingStatefulSet appsv1.StatefulSet
	Eventually(h.createGetKubernetesObjectFunc(eventingStatefulSetName, &eventingStatefulSet)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 10).
		Should(BeTrue())
}

type testHelper struct {
	ctx           context.Context
	namespaceName string
}

func (h *testHelper) getEventingCount() int {
	var objectList v1alpha1.EventingList
	Expect(k8sClient.List(h.ctx, &objectList)).To(Succeed())
	return len(objectList.Items)
}

func (h *testHelper) createGetEventingStateFunc(eventingName string) func() (rtypes.State, error) {
	return func() (rtypes.State, error) {
		return h.getEventingState(eventingName)
	}
}

func (h *testHelper) getEventingState(eventingName string) (rtypes.State, error) {
	var emptyState = rtypes.State("")
	var eventing v1alpha1.Eventing
	key := types.NamespacedName{
		Name:      eventingName,
		Namespace: h.namespaceName,
	}
	err := k8sClient.Get(h.ctx, key, &eventing)
	if err != nil {
		return emptyState, err
	}
	return eventing.Status.State, nil
}

func (h *testHelper) createGetKubernetesObjectFunc(serviceAccountName string, obj client.Object) func() (bool, error) {
	return func() (bool, error) {
		key := types.NamespacedName{
			Name:      serviceAccountName,
			Namespace: h.namespaceName,
		}
		err := k8sClient.Get(h.ctx, key, obj)
		if err != nil {
			return false, err
		}
		return true, err
	}
}

func (h *testHelper) updateStatefulSetStatus(statefulSetName string) {
	By(fmt.Sprintf("Updating statefulSet status: %s", statefulSetName))
	var statefulSet appsv1.StatefulSet
	Eventually(h.createGetKubernetesObjectFunc(statefulSetName, &statefulSet)).
		WithPolling(time.Second * 2).
		WithTimeout(time.Second * 10).
		Should(BeTrue())

	// simulate statefulset status when pods are created
	statefulSet.Status.Replicas = 1
	statefulSet.Status.ReadyReplicas = 1
	statefulSet.Status.CurrentReplicas = 1
	statefulSet.Status.UpdatedReplicas = 1
	statefulSet.Status.AvailableReplicas = 1
	statefulSet.Status.ObservedGeneration = 1
	Expect(k8sClient.Status().Update(h.ctx, &statefulSet)).To(Succeed())

	By(fmt.Sprintf("StatefulSet status updated: %s", statefulSetName))
}

func (h *testHelper) createEventing(eventingName string, spec v1alpha1.EventingSpec) {
	By(fmt.Sprintf("Creating crd: %s", eventingName))
	eventing := v1alpha1.Eventing{
		ObjectMeta: metav1.ObjectMeta{
			Name:      eventingName,
			Namespace: h.namespaceName,
			Labels: map[string]string{
				"operator.kyma-project.io/kyma-name": "test",
			},
		},
		Spec: spec,
	}
	Expect(k8sClient.Create(h.ctx, &eventing)).To(Succeed())
	By(fmt.Sprintf("Crd created: %s", eventingName))
}

func (h *testHelper) createNamespace() {
	By(fmt.Sprintf("Creating namespace: %s", h.namespaceName))
	namespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: h.namespaceName,
		},
	}
	Expect(k8sClient.Create(h.ctx, &namespace)).To(Succeed())
	By(fmt.Sprintf("Namespace created: %s", h.namespaceName))
}
