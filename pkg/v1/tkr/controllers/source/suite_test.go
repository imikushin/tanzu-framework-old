// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package source

import (
	"context"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	capi "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/cluster-api/util/conditions"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	// The fake package is deprecated, though there is talk of undeprecating it
	"sigs.k8s.io/controller-runtime/pkg/client/fake" // nolint:staticcheck
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	// +kubebuilder:scaffold:imports

	runv1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/fakes"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/constants"
)

const (
	version11713 = "v1.17.13---vmware.1"
	version11810 = "v1.18.10---vmware.1"
	version1191  = "v1.19.1---vmware.1"
	version1193  = "v1.19.3---vmware.1"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var bomContent17 []byte
var bomContent18 []byte
var bomContent193 []byte
var bomContent191 []byte
var metadataContent []byte

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "TKR source controller test")
}

func addToScheme(scheme *runtime.Scheme) {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = capi.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	_ = runv1.AddToScheme(scheme)
}

var _ = BeforeSuite(func() {
	bomContent17, _ = os.ReadFile("../../fakes/boms/bom-v1.17.13+vmware.1.yaml")
	bomContent18, _ = os.ReadFile("../../fakes/boms/bom-v1.18.10+vmware.1.yaml")
	bomContent193, _ = os.ReadFile("../../fakes/boms/bom-v1.19.3+vmware.1.yaml")
	bomContent191, _ = os.ReadFile("../../fakes/boms/bom-v1.19.1+vmware.1.yaml")
	metadataContent, _ = os.ReadFile("../../fakes/boms/metadata.yaml")
})

var _ = Describe("SyncRelease", func() {
	var (
		fakeClient   client.Client
		fakeRegistry *fakes.Registry
		scheme       *runtime.Scheme
		objects      []runtime.Object
		r            reconciler
		err          error
	)

	JustBeforeEach(func() {
		scheme = runtime.NewScheme()
		addToScheme(scheme)
		fakeClient = fake.NewFakeClientWithScheme(scheme, objects...)
		r = reconciler{
			client:   fakeClient,
			log:      ctrllog.Log,
			scheme:   scheme,
			registry: fakeRegistry,
			bomImage: "my-registry.io/tkrs",
		}

		err = r.SyncRelease(context.Background())
	})

	Context("When BOM images with proper content are published", func() {

		BeforeEach(func() {
			fakeRegistry = &fakes.Registry{}
			fakeRegistry.ListImageTagsReturns([]string{"bom-v1.17.13+vmware.1", "bom-v1.18.10+vmware.1", "bom-v1.19.3+vmware.1"}, nil)

			// we'll fetch BOM metadata image first
			fakeRegistry.ListImageTagsReturnsOnCall(0, []string{"v1"}, nil)
			fakeRegistry.GetFileReturnsOnCall(0, metadataContent, nil)

			fakeRegistry.GetFileReturnsOnCall(1, bomContent17, nil)
			fakeRegistry.GetFileReturnsOnCall(2, bomContent18, nil)
			fakeRegistry.GetFileReturnsOnCall(3, bomContent193, nil)
			objects = []runtime.Object{}

		})

		It("should create BOM ConfigMaps", func() {
			Expect(err).ToNot(HaveOccurred())

			metadata, err := r.compatibilityMetadata(r.ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(metadata).ToNot(BeNil())

			cmList := &corev1.ConfigMapList{}
			opts := []client.ListOption{
				client.InNamespace(constants.TKRNamespace),
				client.HasLabels{constants.BomConfigMapTKRLabel},
			}
			Expect(fakeClient.List(context.Background(), cmList, opts...)).To(Succeed())
			Expect(len(cmList.Items)).To(Equal(3))
		})

	})

	Context("When a new BOM images is released", func() {

		BeforeEach(func() {
			fakeRegistry = &fakes.Registry{}
			fakeRegistry.ListImageTagsReturns([]string{"bom-v1.17.13+vmware.1", "bom-v1.18.10+vmware.1", "bom-v1.19.3+vmware.1"}, nil)

			// we'll fetch BOM metadata image first
			fakeRegistry.ListImageTagsReturnsOnCall(0, []string{"v1"}, nil)
			fakeRegistry.GetFileReturnsOnCall(0, metadataContent, nil)

			fakeRegistry.GetFileReturnsOnCall(1, bomContent193, nil)

			cm1 := newConfigMap(version11713, map[string]string{constants.BomConfigMapTKRLabel: version11713}, map[string]string{constants.BomConfigMapImageTagAnnotation: "bom-v1.17.13+vmware.1"}, bomContent17)
			cm2 := newConfigMap(version11810, map[string]string{constants.BomConfigMapTKRLabel: version11810}, map[string]string{constants.BomConfigMapImageTagAnnotation: "bom-v1.18.10+vmware.1"}, bomContent18)

			objects = []runtime.Object{cm1, cm2}
		})

		It("should create a new ConfigMap based on the new BOM", func() {
			Expect(err).ToNot(HaveOccurred())
			cmList := &corev1.ConfigMapList{}
			opts := []client.ListOption{
				client.InNamespace(constants.TKRNamespace),
				client.HasLabels{constants.BomConfigMapTKRLabel},
			}
			Expect(fakeClient.List(context.Background(), cmList, opts...)).To(Succeed())
			Expect(len(cmList.Items)).To(Equal(3))
		})
	})

})

var _ = Describe("UpdateTKRCompatibleCondition", func() {
	var (
		tkrs       []runv1.TanzuKubernetesRelease
		fakeClient client.Client
		scheme     *runtime.Scheme
		objects    []runtime.Object
		r          reconciler
		err        error
	)

	JustBeforeEach(func() {
		scheme = runtime.NewScheme()
		addToScheme(scheme)
		fakeClient = fake.NewFakeClientWithScheme(scheme, objects...)
		r = reconciler{
			client:   fakeClient,
			log:      ctrllog.Log,
			scheme:   scheme,
			bomImage: "my-registry.io/tkrs",
		}
		err = r.UpdateTKRCompatibleCondition(context.Background(), tkrs)
	})

	Context("When reconcile the compatible condition of the TKRs", func() {
		BeforeEach(func() {
			tkr1, _ := NewTkrFromBom(version11713, bomContent17)
			tkr2, _ := NewTkrFromBom(version11810, bomContent18)
			tkr3, _ := NewTkrFromBom(version1193, bomContent193)
			tkr4, _ := NewTkrFromBom(version1191, bomContent191)
			cm := newMetadataConfigMap(metadataContent)
			tkrs = []runv1.TanzuKubernetesRelease{tkr1, tkr4, tkr3, tkr2}

			mgmtcluster := newManagementCluster(map[string]string{constants.ManagememtClusterRoleLabel: ""}, map[string]string{constants.TKGVersionKey: "v1.1"})
			objects = []runtime.Object{mgmtcluster, cm}
		})
		It("should update the TKRs' compatible condition", func() {
			Expect(err).ToNot(HaveOccurred())
			for _, tkr := range tkrs {
				if tkr.Name == version1193 {
					status, msg := getConditionStatusAndMessage(tkr.Status.Conditions, runv1.ConditionCompatible)
					Expect(string(status)).To(Equal("False"))
					Expect(msg).To(Equal(""))
				}

				if tkr.Name == version11810 {
					status, msg := getConditionStatusAndMessage(tkr.Status.Conditions, runv1.ConditionCompatible)
					Expect(string(status)).To(Equal("True"))
					Expect(msg).To(Equal(""))
				}

				if tkr.Name == version1191 {
					status, msg := getConditionStatusAndMessage(tkr.Status.Conditions, runv1.ConditionCompatible)
					Expect(string(status)).To(Equal("False"))
					Expect(msg).To(Equal(""))
				}

				if tkr.Name == version11713 {
					status, msg := getConditionStatusAndMessage(tkr.Status.Conditions, runv1.ConditionCompatible)
					Expect(string(status)).To(Equal("True"))
					Expect(msg).To(Equal(""))
				}
			}
		})
	})
})

var _ = Describe("UpdateTKRUpgradeAvailableCondition", func() {
	var (
		tkrs []runv1.TanzuKubernetesRelease
		r    reconciler
	)

	JustBeforeEach(func() {
		r = reconciler{}
		r.UpdateTKRUpdatesAvailableCondition(tkrs)
	})

	Context("When there are available upgrade for some of the TKRs", func() {
		BeforeEach(func() {
			tkr1, _ := NewTkrFromBom(version11713, bomContent17)
			tkr2, _ := NewTkrFromBom(version11810, bomContent18)
			tkr3, _ := NewTkrFromBom(version1193, bomContent193)
			tkr4, _ := NewTkrFromBom(version1191, bomContent191)
			conditions.Set(&tkr2, conditions.TrueCondition(runv1.ConditionCompatible))
			conditions.Set(&tkr3, conditions.TrueCondition(runv1.ConditionCompatible))
			conditions.Set(&tkr4, conditions.TrueCondition(runv1.ConditionCompatible))
			tkrs = []runv1.TanzuKubernetesRelease{tkr1, tkr4, tkr3, tkr2}
		})
		It("should update the UpgradeAvailable Condition with proper message", func() {

			for _, tkr := range tkrs {
				if tkr.Name == version1193 {
					status, msg := getConditionStatusAndMessage(tkr.Status.Conditions, runv1.ConditionUpdatesAvailable)
					Expect(string(status)).To(Equal("False"))
					Expect(msg).To(Equal(""))
				}

				if tkr.Name == version11810 {
					status, msg := getConditionStatusAndMessage(tkr.Status.Conditions, runv1.ConditionUpdatesAvailable)
					Expect(string(status)).To(Equal("True"))
					Expect(msg).To(Equal("[v1.19.1+vmware.1 v1.19.3+vmware.1]"))
				}

				if tkr.Name == version1191 {
					status, msg := getConditionStatusAndMessage(tkr.Status.Conditions, runv1.ConditionUpdatesAvailable)
					Expect(string(status)).To(Equal("True"))
					Expect(msg).To(Equal("[v1.19.3+vmware.1]"))
				}

				if tkr.Name == version11713 {
					status, msg := getConditionStatusAndMessage(tkr.Status.Conditions, runv1.ConditionUpdatesAvailable)
					Expect(string(status)).To(Equal("True"))
					Expect(msg).To(Equal("[v1.18.10+vmware.1]"))
				}
			}
		})
	})
})

var _ = Describe("initialReconcile", func() {
	var (
		fakeClient   client.Client
		fakeRegistry *fakes.Registry
		scheme       *runtime.Scheme
		objects      []runtime.Object
		r            reconciler
		stopChan     chan struct{}
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()
		addToScheme(scheme)
		fakeClient = fake.NewFakeClientWithScheme(scheme, objects...)
	})
	JustBeforeEach(func() {
		r = reconciler{
			client:                     fakeClient,
			log:                        ctrllog.Log,
			scheme:                     scheme,
			registry:                   fakeRegistry,
			bomImage:                   "my-registry.io/tkrs",
			compatibilityMetadataImage: "",
		}
		ticker := time.NewTicker(time.Second)
		stopChan = make(chan struct{})
		go func(stopChan chan struct{}) {
			time.Sleep(time.Second * 5)
			stopChan <- struct{}{}
		}(stopChan)
		r.initialReconcile(ticker, stopChan, 3)
	})

	Context("When in initial sync-up stage", func() {
		BeforeEach(func() {
			fakeRegistry = &fakes.Registry{}
			fakeRegistry.ListImageTagsReturns([]string{"bom-v1.17.13+vmware.1", "bom-v1.18.10+vmware.1", "bom-v1.19.3+vmware.1"}, nil)

			// we'll fetch BOM metadata image first
			fakeRegistry.ListImageTagsReturnsOnCall(0, []string{"v1"}, nil)
			fakeRegistry.GetFileReturnsOnCall(0, metadataContent, nil)

			fakeRegistry.GetFileReturnsOnCall(1, bomContent17, nil)
			fakeRegistry.GetFileReturnsOnCall(2, bomContent18, nil)
			fakeRegistry.GetFileReturnsOnCall(3, bomContent193, nil)
		})
		It("retrieve the list of images at least once", func() {
			Expect(fakeRegistry.ListImageTagsCallCount()).Should(BeNumerically(">=", 1))
		})
	})

	Context("When cluster is ready, but the some BOM content can not be retrieved", func() {
		BeforeEach(func() {
			fakeRegistry = &fakes.Registry{}
			fakeRegistry.ListImageTagsReturns([]string{"bom-v1.17.13+vmware.1", "bom-v1.18.10+vmware.1", "bom-v1.19.3+vmware.1"}, nil)

			// we'll fetch BOM metadata image first
			fakeRegistry.ListImageTagsReturnsOnCall(0, []string{"v1"}, nil)
			fakeRegistry.GetFileReturnsOnCall(0, metadataContent, nil)

			fakeRegistry.GetFileReturnsOnCall(1, bomContent17, nil)
			fakeRegistry.GetFileReturnsOnCall(2, nil, errors.New("fake-error"))
			fakeRegistry.GetFileReturnsOnCall(3, bomContent193, nil)
			mgmtcluster := newManagementCluster(map[string]string{constants.ManagememtClusterRoleLabel: ""}, map[string]string{constants.TKGVersionKey: "v1.1"})
			objects = []runtime.Object{mgmtcluster}
			fakeClient = fake.NewFakeClientWithScheme(scheme, objects...)
		})

		It("should retrieve what can be retrieved and create appropriate ConfigMaps", func() {
			cmList := &corev1.ConfigMapList{}
			opts := []client.ListOption{
				client.InNamespace(constants.TKRNamespace),
				client.HasLabels{constants.BomConfigMapTKRLabel},
			}
			Expect(fakeClient.List(context.Background(), cmList, opts...)).To(Succeed())
			Expect(len(cmList.Items)).To(Equal(2))

			metadata, err := r.compatibilityMetadata(r.ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(metadata).ToNot(BeNil())
		})
	})

	Context("When cluster is ready, and bom content can be retrieved", func() {
		BeforeEach(func() {
			fakeRegistry = &fakes.Registry{}
			fakeRegistry.ListImageTagsReturns([]string{"bom-v1.17.13+vmware.1", "bom-v1.18.10+vmware.1", "bom-v1.19.3+vmware.1"}, nil)

			// we'll fetch BOM metadata image first
			fakeRegistry.ListImageTagsReturnsOnCall(0, []string{"v1"}, nil)
			fakeRegistry.GetFileReturnsOnCall(0, metadataContent, nil)

			fakeRegistry.GetFileReturnsOnCall(1, bomContent17, nil)
			fakeRegistry.GetFileReturnsOnCall(2, bomContent18, nil)
			fakeRegistry.GetFileReturnsOnCall(3, bomContent193, nil)

			mgmtcluster := newManagementCluster(map[string]string{constants.ManagememtClusterRoleLabel: ""}, map[string]string{constants.TKGVersionKey: "v1.1"})
			objects = []runtime.Object{mgmtcluster}
			fakeClient = fake.NewFakeClientWithScheme(scheme, objects...)
		})

		It("should create the metadata ConfigMap and all BOM ConfigMaps", func() {
			cmList := &corev1.ConfigMapList{}
			opts := []client.ListOption{
				client.InNamespace(constants.TKRNamespace),
				client.HasLabels{constants.BomConfigMapTKRLabel},
			}
			Expect(fakeClient.List(context.Background(), cmList, opts...)).To(Succeed())
			Expect(len(cmList.Items)).To(Equal(3))

			metadata, err := r.compatibilityMetadata(r.ctx)
			Expect(err).ToNot(HaveOccurred())
			Expect(metadata).ToNot(BeNil())
		})

		When("creating a ConfigMap returns an error", func() {
			BeforeEach(func() {
				fakeClient = clientErrOnCreate{Client: fakeClient, err: errors.New("EXPECTED"), errOnName: "v1.18.10---vmware.1"}
				fakeClient = clientErrOnCreate{Client: fakeClient, err: errors.New("EXPECTED"), errOnName: constants.BOMMetadataConfigMapName}
			})

			It("should create the metadata ConfigMap and all BOM ConfigMaps except those affected by the error", func() {
				cmList := &corev1.ConfigMapList{}
				opts := []client.ListOption{
					client.InNamespace(constants.TKRNamespace),
					client.HasLabels{constants.BomConfigMapTKRLabel},
				}
				Expect(fakeClient.List(context.Background(), cmList, opts...)).To(Succeed())
				Expect(len(cmList.Items)).To(Equal(2)) // not all 3 are expected

				metadata, err := r.compatibilityMetadata(r.ctx)
				Expect(err).To(HaveOccurred())
				Expect(metadata).To(BeNil())
			})
		})
	})
})

type clientErrOnCreate struct {
	client.Client
	err       error
	errOnName string
}

func (c clientErrOnCreate) Create(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) error {
	if cm, ok := obj.(*corev1.ConfigMap); ok && cm.Name == c.errOnName {
		return c.err
	}
	return c.Client.Create(ctx, obj, opts...)
}

var _ = Describe("r.Reconcile()", func() {
	var (
		fakeRegistry *fakes.Registry
		fakeClient   client.Client
		scheme       *runtime.Scheme
		objects      []runtime.Object
		r            reconciler
	)

	JustBeforeEach(func() {
		scheme = runtime.NewScheme()
		addToScheme(scheme)
		fakeClient = fake.NewFakeClientWithScheme(scheme, objects...)
		r = reconciler{
			registry:                   fakeRegistry,
			ctx:                        context.Background(),
			client:                     fakeClient,
			log:                        ctrllog.Log,
			scheme:                     scheme,
			compatibilityMetadataImage: "",
		}
	})

	When("new BOM ConfigMaps are added", func() {
		var (
			cm1, cm2 *corev1.ConfigMap
			err      error
		)

		BeforeEach(func() {
			cm1 = newConfigMap(version11713, map[string]string{constants.BomConfigMapTKRLabel: version11713}, map[string]string{constants.BomConfigMapImageTagAnnotation: "bom-v1.17.13+vmware.1"}, bomContent17)
			cm2 = newConfigMap(version1193, map[string]string{constants.BomConfigMapTKRLabel: version1193}, map[string]string{constants.BomConfigMapImageTagAnnotation: "bom-v1.19.3+vmware.1"}, bomContent193)
			tkr1, _ := NewTkrFromBom(version11713, bomContent17)
			mgmtCluster := newManagementCluster(map[string]string{constants.ManagememtClusterRoleLabel: ""}, map[string]string{constants.TKGVersionKey: "v1.1"})
			cmMeta := newMetadataConfigMap(metadataContent)

			objects = []runtime.Object{mgmtCluster, cmMeta, cm1, cm2, &tkr1}
		})

		It("should create the corresponding TKRs", func() {
			_, err = r.Reconcile(req(cm2))
			Expect(err).ToNot(HaveOccurred())

			tkrList := &runv1.TanzuKubernetesReleaseList{}
			Expect(r.client.List(r.ctx, tkrList)).To(Succeed())
			Expect(tkrList.Items).To(HaveLen(2))

			found1193 := false
			for i := range tkrList.Items {
				tkr := &tkrList.Items[i]
				if tkr.Name == "v1.19.3---vmware.1" {
					found1193 = true
					status, _ := getConditionStatusAndMessage(tkr.Status.Conditions, runv1.ConditionCompatible)
					Expect(status).To(Equal(corev1.ConditionFalse))
					continue
				}
				status, _ := getConditionStatusAndMessage(tkr.Status.Conditions, runv1.ConditionCompatible)
				Expect(status).To(Equal(corev1.ConditionTrue))
			}
			Expect(found1193).To(BeTrue())

		})
	})

	When("a TKR already exists for the BOM ConfigMap", func() {
		var (
			cm1, cm2 *corev1.ConfigMap
			err      error
		)

		BeforeEach(func() {
			cm1 = newConfigMap(version11713, map[string]string{constants.BomConfigMapTKRLabel: version11713}, map[string]string{constants.BomConfigMapImageTagAnnotation: "bom-v1.17.13+vmware.1"}, bomContent17)
			cm2 = newConfigMap(version1193, map[string]string{constants.BomConfigMapTKRLabel: version1193}, map[string]string{constants.BomConfigMapImageTagAnnotation: "bom-v1.19.3+vmware.1"}, bomContent193)
			tkr1, _ := NewTkrFromBom(version11713, bomContent17)
			mgmtCluster := newManagementCluster(map[string]string{constants.ManagememtClusterRoleLabel: ""}, map[string]string{constants.TKGVersionKey: "v1.1"})
			cmMeta := newMetadataConfigMap(metadataContent)

			objects = []runtime.Object{mgmtCluster, cmMeta, cm1, cm2, &tkr1}
		})

		It("should not return an error", func() {
			_, err = r.Reconcile(req(cm1))
			Expect(err).ToNot(HaveOccurred())

			tkrList := &runv1.TanzuKubernetesReleaseList{}
			Expect(r.client.List(r.ctx, tkrList)).To(Succeed())
			Expect(tkrList.Items).To(HaveLen(1))
		})
	})

	When("a TKR cannot be created for the BOM ConfigMap", func() {
		var (
			cm1, cm2 *corev1.ConfigMap
			err      error
		)

		BeforeEach(func() {
			cm1 = newConfigMap(version11713, map[string]string{constants.BomConfigMapTKRLabel: version11713}, map[string]string{constants.BomConfigMapImageTagAnnotation: "bom-v1.17.13+vmware.1"}, bomContent17)
			cm2 = newConfigMap(version1193, map[string]string{constants.BomConfigMapTKRLabel: version1193}, map[string]string{constants.BomConfigMapImageTagAnnotation: "bom-v1.19.3+vmware.1"}, bomContent193)
			mgmtCluster := newManagementCluster(map[string]string{constants.ManagememtClusterRoleLabel: ""}, map[string]string{constants.TKGVersionKey: "v1.1"})
			cmMeta := newMetadataConfigMap(metadataContent)

			delete(cm1.BinaryData, constants.BomConfigMapContentKey)
			delete(cm2.Labels, constants.BomConfigMapTKRLabel)

			objects = []runtime.Object{mgmtCluster, cmMeta, cm1, cm2}
		})

		It("should not return an error, and TKR is not created", func() {
			_, err = r.Reconcile(req(cm1))
			Expect(err).ToNot(HaveOccurred())
			_, err = r.Reconcile(req(cm2))
			Expect(err).ToNot(HaveOccurred())

			tkrList := &runv1.TanzuKubernetesReleaseList{}
			Expect(r.client.List(r.ctx, tkrList)).To(Succeed())
			Expect(tkrList.Items).To(HaveLen(0)) // no TKRs created
		})
	})

	When("reconciling the BOM metadata ConfigMap", func() {
		var (
			cm1, cmMeta *corev1.ConfigMap
			err         error
		)

		BeforeEach(func() {
			cm1 = newConfigMap(version11713, map[string]string{constants.BomConfigMapTKRLabel: version11713}, map[string]string{constants.BomConfigMapImageTagAnnotation: "bom-v1.17.13+vmware.1"}, bomContent17)
			tkr1, _ := NewTkrFromBom(version11713, bomContent17)
			mgmtCluster := newManagementCluster(map[string]string{constants.ManagememtClusterRoleLabel: ""}, map[string]string{constants.TKGVersionKey: "v1.1"})
			cmMeta = newMetadataConfigMap(metadataContent)

			objects = []runtime.Object{mgmtCluster, cmMeta, cm1, &tkr1}
		})

		When("management cluster is in place", func() {
			It("should not return an error", func() {
				_, err = r.Reconcile(req(cmMeta))
				Expect(err).ToNot(HaveOccurred())

				tkrList := &runv1.TanzuKubernetesReleaseList{}
				Expect(r.client.List(r.ctx, tkrList)).To(Succeed())
				Expect(tkrList.Items).To(HaveLen(1))
			})
		})

		When("there's an error getting the management cluster info", func() {
			expectedErr := errors.New("this is a bad day for clusters")

			JustBeforeEach(func() {
				r.client = clientErrOnGetCluster{Client: r.client, err: expectedErr}
			})

			It("should return that error, so reconciliation would be retried", func() {
				_, err := r.Reconcile(req(cmMeta))
				Expect(err).To(HaveOccurred())
				Expect(errors.Cause(err)).To(Equal(expectedErr))
			})
		})
	})

	When("BOM metadata ConfigMap cannot be obtained", func() {
		var (
			cm1, cm2 *corev1.ConfigMap
			err      error
		)

		BeforeEach(func() {
			cm1 = newConfigMap(version11713, map[string]string{constants.BomConfigMapTKRLabel: version11713}, map[string]string{constants.BomConfigMapImageTagAnnotation: "bom-v1.17.13+vmware.1"}, bomContent17)
			cm2 = newConfigMap(version1193, map[string]string{constants.BomConfigMapTKRLabel: version1193}, map[string]string{constants.BomConfigMapImageTagAnnotation: "bom-v1.19.3+vmware.1"}, bomContent193)
			tkr1, _ := NewTkrFromBom(version11713, bomContent17)
			mgmtCluster := newManagementCluster(map[string]string{constants.ManagememtClusterRoleLabel: ""}, map[string]string{constants.TKGVersionKey: "v1.1"})

			objects = []runtime.Object{mgmtCluster, cm1, cm2, &tkr1}
		})

		It("should still create the TKRs, but with default status conditions", func() {
			_, err = r.Reconcile(req(cm2))
			Expect(err).To(HaveOccurred())
			Expect(apierrors.IsNotFound(errors.Cause(err))).To(BeTrue())

			tkrList := &runv1.TanzuKubernetesReleaseList{}
			Expect(r.client.List(r.ctx, tkrList)).To(Succeed())
			Expect(tkrList.Items).To(HaveLen(2))

			for i := range tkrList.Items {
				tkr := &tkrList.Items[i]
				condition := conditions.Get(tkr, runv1.ConditionCompatible)
				Expect(condition == nil || condition.Status == corev1.ConditionUnknown || condition.Status == corev1.ConditionFalse).To(BeTrue())
			}
		})
	})
})

type clientErrOnGetCluster struct {
	client.Client
	err error
}

func (c clientErrOnGetCluster) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
	if _, ok := obj.(*capi.Cluster); !ok {
		return c.err
	}
	return c.Client.Get(ctx, key, obj)
}

var _ = Describe("errorSlice.Error()", func() {
	err := errorSlice{errors.New("one"), errors.New("two"), errors.New("three")}
	Expect(err.Error()).To(Equal("one, two, three"))
	Expect(errorSlice{}.Error()).To(Equal(""))
})

func req(o metav1.Object) ctrl.Request {
	return ctrl.Request{NamespacedName: client.ObjectKey{Namespace: o.GetNamespace(), Name: o.GetName()}}
}

func getConditionStatusAndMessage(conditions []capi.Condition, conditionType capi.ConditionType) (status corev1.ConditionStatus, msg string) {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			status = condition.Status
			msg = condition.Message
			return
		}
	}
	status = corev1.ConditionStatus("")
	return
}

func newConfigMap(name string, labels, annotations map[string]string, content []byte) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   constants.TKRNamespace,
			Labels:      labels,
			Annotations: annotations,
		},
		BinaryData: map[string][]byte{constants.BomConfigMapContentKey: content},
	}
}

func newMetadataConfigMap(content []byte) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.BOMMetadataConfigMapName,
			Namespace: constants.TKRNamespace,
		},
		BinaryData: map[string][]byte{constants.BOMMetadataCompatibilityKey: content},
	}
}

func newManagementCluster(labels, annotations map[string]string) *capi.Cluster {
	return &capi.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "mgmt-cluster",
			Namespace:   constants.TKGNamespace,
			Labels:      labels,
			Annotations: annotations,
		},
	}
}
