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

var _ = Describe("ProxmoxMachineReconciler", Label("unit", "controllers"), func() {
	BeforeEach(func() {})
	AfterEach(func() {})

	Context("Reconcile an ProxmoxMachine", func() {
		It("should not error with minimal set up", func() {
			reconciler := &ProxmoxMachineReconciler{
				Client: k8sClient,
			}
			By("Calling reconcile")
			ctx := context.Background()
			instance := &infrav1.ProxmoxMachine{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "default"}}
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
