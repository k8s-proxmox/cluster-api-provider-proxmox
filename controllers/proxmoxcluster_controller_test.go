package controllers

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
)

var _ = Describe("ProxmoxClusterReconciler", func() {
	BeforeEach(func() {})
	AfterEach(func() {})

	Context("Reconcile ProxmoxCluster", func() {
		It("should not error and requeue the request with insufficient set up", func() {
			ctx := context.Background()

			reconciler := &ProxmoxClusterReconciler{
				Client: k8sClient,
			}

			instance := &infrav1.ProxmoxCluster{
				ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "default"},
				Spec: infrav1.ProxmoxClusterSpec{
					ServerRef: infrav1.ServerRef{
						Endpoint:  "a.b.c.d:8006",
						SecretRef: &infrav1.ObjectReference{Name: "foo"},
					},
				},
			}

			// Create the ProxmoxCluster object and expect the Reconcile and Deployment to be created
			Expect(k8sClient.Create(ctx, instance)).To(Succeed())
			defer func() {
				err := k8sClient.Delete(ctx, instance)
				Expect(err).NotTo(HaveOccurred())
			}()

			result, err := reconciler.Reconcile(ctx, ctrl.Request{
				NamespacedName: client.ObjectKey{
					Namespace: instance.Namespace,
					Name:      instance.Name,
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.RequeueAfter).To(BeZero())
			Expect(result.Requeue).To(BeFalse())
		})
	})
})
