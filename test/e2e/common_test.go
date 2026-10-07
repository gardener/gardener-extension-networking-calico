// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package e2e_test

import (
	"context"
	"os"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	"github.com/gardener/gardener/test/framework"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/utils/ptr"

	calicoinstall "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/install"
	calicov1alpha1 "github.com/gardener/gardener-extension-networking-calico/pkg/apis/calico/v1alpha1"
)

var (
	parentCtx context.Context
	encoder   runtime.Encoder
)

func init() {
	scheme := runtime.NewScheme()
	utilruntime.Must(calicoinstall.AddToScheme(scheme))
	encoder = serializer.NewCodecFactory(scheme).LegacyCodec(calicov1alpha1.SchemeGroupVersion)
}

var _ = BeforeEach(func() {
	parentCtx = context.Background()
})

const projectNamespace = "garden-local"

func defaultShootCreationFramework() *framework.ShootCreationFramework {
	kubeconfigPath := os.Getenv("KUBECONFIG")
	return framework.NewShootCreationFramework(&framework.ShootCreationConfig{
		GardenerConfig: &framework.GardenerConfig{
			ProjectNamespace:   projectNamespace,
			GardenerKubeconfig: kubeconfigPath,
			SkipAccessingShoot: false,
			CommonConfig:       &framework.CommonConfig{},
		},
	})
}

func defaultShoot(generateName string) *gardencorev1beta1.Shoot {
	return &gardencorev1beta1.Shoot{
		ObjectMeta: metav1.ObjectMeta{
			Name: generateName,
			Annotations: map[string]string{
				v1beta1constants.AnnotationShootCloudConfigExecutionMaxDelaySeconds: "0",
			},
		},
		Spec: gardencorev1beta1.ShootSpec{
			Region:                 "local",
			CredentialsBindingName: ptr.To("local"),
			CloudProfile: &gardencorev1beta1.CloudProfileReference{
				Name: "local",
				Kind: "CloudProfile",
			},
			Kubernetes: gardencorev1beta1.Kubernetes{
				Version: "1.33.0",
				Kubelet: &gardencorev1beta1.KubeletConfig{
					SerializeImagePulls: ptr.To(false),
					RegistryPullQPS:     ptr.To[int32](10),
					RegistryBurst:       ptr.To[int32](20),
				},
				KubeAPIServer: &gardencorev1beta1.KubeAPIServerConfig{},
			},
			Networking: &gardencorev1beta1.Networking{
				Type:       ptr.To("calico"),
				IPFamilies: []gardencorev1beta1.IPFamily{gardencorev1beta1.IPFamilyIPv4},
				Nodes:      ptr.To("10.0.0.0/16"),
			},
			Provider: gardencorev1beta1.Provider{
				Type: "local",
				Workers: []gardencorev1beta1.Worker{{
					Name: "local",
					Machine: gardencorev1beta1.Machine{
						Type: "local",
					},
					CRI: &gardencorev1beta1.CRI{
						Name: gardencorev1beta1.CRINameContainerD,
					},
					Minimum: 2,
					Maximum: 2,
				}},
			},
			SystemComponents: &gardencorev1beta1.SystemComponents{
				NodeLocalDNS: &gardencorev1beta1.NodeLocalDNS{
					Enabled: true,
				},
			},
		},
	}
}

func ebpfShoot(generateName string) *gardencorev1beta1.Shoot {
	GinkgoHelper()
	shoot := defaultShoot(generateName)
	shoot.Spec.Kubernetes.KubeProxy = &gardencorev1beta1.KubeProxyConfig{
		Enabled: new(false),
	}

	networkConfig := &calicov1alpha1.NetworkConfig{
		EbpfDataplane: &calicov1alpha1.EbpfDataplane{
			Enabled: true,
		},
	}
	networkConfigRaw, err := runtime.Encode(encoder, networkConfig)
	Expect(err).NotTo(HaveOccurred())
	shoot.Spec.Networking.ProviderConfig = &runtime.RawExtension{
		Raw: networkConfigRaw,
	}
	return shoot
}
