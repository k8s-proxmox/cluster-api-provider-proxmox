package scope

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	infrav1 "github.com/k8s-proxmox/cluster-api-provider-proxmox/api/v1beta1"
)

var _ = Describe("newComputeService", Label("unit", "scope"), func() {
	var cluster *infrav1.ProxmoxCluster

	Context("When SecretRef is not present in ProxmoxCluster", func() {
		BeforeEach(func() {
			cluster = &infrav1.ProxmoxCluster{}
		})
		It("should return proper error", func() {
			svc, err := newComputeService(context.TODO(), cluster, k8sClient)
			Expect(err.Error()).To(Equal("failed to get proxmox client from nil secretRef"))
			Expect(svc).To(BeNil())
		})
	})

	Context("When Secret is not present", func() {
		BeforeEach(func() {
			cluster = &infrav1.ProxmoxCluster{}
			cluster.Spec.ServerRef.SecretRef = &infrav1.ObjectReference{
				Namespace: "default",
				Name:      "foo",
			}
		})

		It("should return proper error", func() {
			svc, err := newComputeService(context.TODO(), cluster, k8sClient)
			Expect(errors.IsNotFound(err)).To(BeTrue())
			Expect(svc).To(BeNil())
		})
	})

	Context("When Secret is empty", func() {
		BeforeEach(func() {
			cluster = &infrav1.ProxmoxCluster{}
			cluster.Spec.ServerRef.SecretRef = &infrav1.ObjectReference{
				Namespace: "default",
				Name:      "foo",
			}
			cluster.SetName("foo")
			cluster.SetUID("bar")
			secret := &corev1.Secret{}
			secret.SetNamespace("default")
			secret.SetName("foo")
			err := k8sClient.Create(context.TODO(), secret)
			Expect(err).To(BeNil())
		})

		It("Should return proper error", func() {
			svc, err := newComputeService(context.TODO(), cluster, k8sClient)
			Expect(err.Error()).To(Equal("invalid authentication config"))
			Expect(svc).To(BeNil())
		})
	})
})
