/*
Copyright 2023 Simplysoft GmbH.

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

package instance

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	infrav1 "github.com/sp-yduck/cluster-api-provider-proxmox/api/v1beta1"
	"github.com/sp-yduck/cluster-api-provider-proxmox/cloud/services"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	caipamv1 "sigs.k8s.io/cluster-api/exp/ipam/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
	"time"
)

// generates the primary network config. If IP pools are configured, ensures to claim an address from the pool,
// otherwise falls back to configured ip addresses on the IPConfig or DHCP if none are provided
func (s *Service) generateIpConfig0(ctx context.Context) (string, error) {
	template := s.scope.GetProxmoxMachineTemplate(ctx)
	machine := *s.scope.GetProxmoxMachine()
	k8sClient := *s.scope.K8sClient()

	requeue := false
	var ipv4Addresses *addressFromPool
	var ipv6Addresses *addressFromPool
	if template.Spec.Template.Spec.Network.IPConfig.IPv4FromPoolRef != nil && template.Spec.Template.Spec.Network.IPConfig.IPv4FromPoolRef.Name != "" {
		ref := getRef(template.Spec.Template.Spec.Network.IPConfig.IPv4FromPoolRef)

		rc, err := ensureIPClaim(ctx, k8sClient, machine, *ref)
		if err != nil {
			return "", err
		}
		var itemRequeue bool
		addr, itemRequeue, err := addressFromClaim(ctx, k8sClient, machine, *ref, rc.claim)
		requeue = requeue || itemRequeue
		if err != nil {
			return "", err
		} else {
			ipv4Addresses = &addr
		}
	}

	if template.Spec.Template.Spec.Network.IPConfig.IPv6FromPoolRef != nil && template.Spec.Template.Spec.Network.IPConfig.IPv6FromPoolRef.Name != "" {
		ref := getRef(template.Spec.Template.Spec.Network.IPConfig.IPv6FromPoolRef)

		rc, err := ensureIPClaim(ctx, k8sClient, machine, *ref)
		if err != nil {
			return "", err
		}
		var itemRequeue bool
		addr, itemRequeue, err := addressFromClaim(ctx, k8sClient, machine, *ref, rc.claim)
		requeue = requeue || itemRequeue
		if err != nil {
			return "", err
		} else {
			ipv6Addresses = &addr
		}
	}

	if requeue {
		return "", services.WithTransientError(fmt.Errorf("not all ip addresses available"), time.Second*5)
	}

	var configs []string
	if ipv4Addresses != nil {
		configs = append(configs, fmt.Sprintf("ip=%s/%d", ipv4Addresses.Address, ipv4Addresses.Prefix))
		if ipv4Addresses.Gateway != "" {
			configs = append(configs, fmt.Sprintf("gw=%s", ipv4Addresses.Gateway))
		}
	}
	if ipv6Addresses != nil {
		configs = append(configs, fmt.Sprintf("ip6=%s/%d", ipv6Addresses.Address, ipv6Addresses.Prefix))
		if ipv6Addresses.Gateway != "" {
			configs = append(configs, fmt.Sprintf("gw6=%s", ipv6Addresses.Gateway))
		}
	}

	if len(configs) > 0 {
		return strings.Join(configs, ","), nil
	} else {
		return machine.Spec.Network.IPConfig.String(), nil
	}
}

func getRef(ref *corev1.TypedLocalObjectReference) *corev1.TypedLocalObjectReference {
	if ref.APIGroup == nil || *ref.APIGroup == "" {
		ref.APIGroup = pointer.String("ipam.cluster.x-k8s.io")
	}

	return ref
}

type addressFromPool struct {
	Address string
	Prefix  int
	Gateway string
	//dnsServers []string
}

type reconciledClaim struct {
	claim      *caipamv1.IPAddressClaim
	fetchAgain bool
}

func ensureIPClaim(ctx context.Context, client client.Client, m infrav1.ProxmoxMachine, poolRef corev1.TypedLocalObjectReference) (reconciledClaim, error) {
	claim := &caipamv1.IPAddressClaim{}
	nn := types.NamespacedName{
		Namespace: m.Namespace,
		Name:      m.Name + "-" + poolRef.Name,
	}

	if err := client.Get(ctx, nn, claim); err != nil {
		if !apierrors.IsNotFound(err) {
			return reconciledClaim{claim: claim}, err
		}
	}
	if claim.Name != "" {
		return reconciledClaim{claim: claim}, nil
	}

	// No claim exists, we create a new one
	claim = &caipamv1.IPAddressClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      addressClaimName(m, poolRef),
			Namespace: m.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: m.APIVersion,
					Kind:       m.Kind,
					Name:       m.Name,
					UID:        m.UID,
					Controller: pointer.BoolPtr(true),
				},
			},
			Labels: m.Labels,
			/*Finalizers: []string{
				infrav1.MachineFinalizerIPAddressClaim,
			},*/
		},
		Spec: caipamv1.IPAddressClaimSpec{
			PoolRef: poolRef,
		},
	}

	err := client.Create(ctx, claim)
	// if the claim already exists we can try to fetch it again
	if err == nil || apierrors.IsAlreadyExists(err) {
		return reconciledClaim{claim: claim, fetchAgain: true}, nil
	}
	return reconciledClaim{claim: claim}, err
}

func releaseAddressFromPool(ctx context.Context, client client.Client, m infrav1.ProxmoxMachine, poolRef corev1.TypedLocalObjectReference) error {
	claim := &caipamv1.IPAddressClaim{}
	nn := types.NamespacedName{
		Namespace: m.Namespace,
		Name:      addressClaimName(m, poolRef),
	}
	if err := client.Get(ctx, nn, claim); err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		return nil
	}

	if controllerutil.RemoveFinalizer(claim, infrav1.MachineFinalizerIPAddressClaim) {
		if err := client.Update(ctx, claim); err != nil {
			return err
		}
	}

	err := client.Delete(ctx, claim)
	if apierrors.IsNotFound(err) {
		return nil
	}
	return err
}

func addressClaimName(m infrav1.ProxmoxMachine, poolRef corev1.TypedLocalObjectReference) string {
	return m.Name + "-" + poolRef.Name
}

// addressFromClaim retrieves the IPAddress for a CAPI IPAddressClaim.
func addressFromClaim(ctx context.Context, client client.Client, m infrav1.ProxmoxMachine, _ corev1.TypedLocalObjectReference, claim *caipamv1.IPAddressClaim) (addressFromPool, bool, error) {
	log := log.FromContext(ctx)

	if claim == nil {
		return addressFromPool{}, true, errors.New("no claim provided")
	}
	if !claim.DeletionTimestamp.IsZero() {
		// This IPClaim is about to be deleted, so we cannot use it. Requeue.
		log.Info("Found IPClaim with deletion timestamp, requeuing.", "IPClaim", claim)
		return addressFromPool{}, true, nil
	}

	if claim.Status.AddressRef.Name == "" {
		return addressFromPool{}, true, nil
	}

	address := &caipamv1.IPAddress{}
	addressNamespacedName := types.NamespacedName{
		Name:      claim.Status.AddressRef.Name,
		Namespace: m.Namespace,
	}

	if err := client.Get(ctx, addressNamespacedName, address); err != nil {
		if apierrors.IsNotFound(err) {
			return addressFromPool{}, true, nil
		}
		return addressFromPool{}, false, err
	}

	a := addressFromPool{
		Address: address.Spec.Address,
		Prefix:  address.Spec.Prefix,
		Gateway: address.Spec.Gateway,
	}
	log.Info("allocating", "addr", a)
	return a, false, nil
}
