package config

import (
	"context"
	"testing"

	"github.com/redhat-appstudio/helmet/internal/annotations"
	"github.com/redhat-appstudio/helmet/internal/constants"
	"github.com/redhat-appstudio/helmet/internal/k8s"

	o "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestNewConfigMapManager tests the constructor
func TestNewConfigMapManager(t *testing.T) {
	g := o.NewWithT(t)

	fakeKube := k8s.NewFakeKube()
	mgr := NewConfigMapManager(fakeKube, "test-app")

	g.Expect(mgr).ToNot(o.BeNil())
	g.Expect(mgr.Name()).To(o.Equal("test-app-config"))
}

// TestGetConfigMap tests retrieval of ConfigMap from cluster
func TestGetConfigMap(t *testing.T) {
	ctx := context.Background()

	t.Run("success_single_configmap", func(t *testing.T) {
		g := o.NewWithT(t)

		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-config",
				Namespace: "test-ns",
				Labels: map[string]string{
					annotations.Config: "true",
				},
			},
			Data: map[string]string{
				constants.ConfigFilename: "test: config",
			},
		}

		fakeKube := k8s.NewFakeKube(cm)
		mgr := NewConfigMapManager(fakeKube, "test-app")

		result, err := mgr.GetConfigMap(ctx)
		g.Expect(err).To(o.Succeed())
		g.Expect(result).ToNot(o.BeNil())
		g.Expect(result.Name).To(o.Equal("test-config"))
		g.Expect(result.Namespace).To(o.Equal("test-ns"))
	})

	t.Run("error_no_configmap", func(t *testing.T) {
		g := o.NewWithT(t)

		fakeKube := k8s.NewFakeKube()
		mgr := NewConfigMapManager(fakeKube, "test-app")

		result, err := mgr.GetConfigMap(ctx)
		g.Expect(err).To(o.HaveOccurred())
		g.Expect(err).To(o.MatchError(o.ContainSubstring("cluster configmap not found")))
		g.Expect(result).To(o.BeNil())
	})

	t.Run("error_multiple_configmaps", func(t *testing.T) {
		g := o.NewWithT(t)

		cm1 := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "config-1",
				Namespace: "ns1",
				Labels: map[string]string{
					annotations.Config: "true",
				},
			},
		}
		cm2 := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "config-2",
				Namespace: "ns2",
				Labels: map[string]string{
					annotations.Config: "true",
				},
			},
		}

		fakeKube := k8s.NewFakeKube(cm1, cm2)
		mgr := NewConfigMapManager(fakeKube, "test-app")

		result, err := mgr.GetConfigMap(ctx)
		g.Expect(err).To(o.HaveOccurred())
		g.Expect(err).To(o.MatchError(o.ContainSubstring("multiple cluster configmaps found")))
		g.Expect(err).To(o.MatchError(o.ContainSubstring("ns1/config-1")))
		g.Expect(err).To(o.MatchError(o.ContainSubstring("ns2/config-2")))
		g.Expect(result).To(o.BeNil())
	})
}

// TestGetConfig tests retrieving and parsing configuration from ConfigMap
func TestGetConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		g := o.NewWithT(t)

		validConfig := `tssc:
  namespace: test-namespace
  settings:
    crc: false
  products: []
`

		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-config",
				Namespace: "test-ns",
				Labels: map[string]string{
					annotations.Config: "true",
				},
			},
			Data: map[string]string{
				constants.ConfigFilename: validConfig,
			},
		}

		fakeKube := k8s.NewFakeKube(cm)
		mgr := NewConfigMapManager(fakeKube, "test-app")

		cfg, err := mgr.GetConfig(ctx)
		g.Expect(err).To(o.Succeed())
		g.Expect(cfg).ToNot(o.BeNil())
		g.Expect(cfg.Namespace()).To(o.Equal("test-ns"))
	})

	t.Run("error_missing_data_key", func(t *testing.T) {
		g := o.NewWithT(t)

		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-config",
				Namespace: "test-ns",
				Labels: map[string]string{
					annotations.Config: "true",
				},
			},
			Data: map[string]string{
				"wrong-key": "value",
			},
		}

		fakeKube := k8s.NewFakeKube(cm)
		mgr := NewConfigMapManager(fakeKube, "test-app")

		cfg, err := mgr.GetConfig(ctx)
		g.Expect(err).To(o.HaveOccurred())
		g.Expect(err).To(o.MatchError(o.ContainSubstring("invalid configmap found in the cluster")))
		g.Expect(err).To(o.MatchError(o.ContainSubstring(constants.ConfigFilename)))
		g.Expect(cfg).To(o.BeNil())
	})

	t.Run("error_empty_data", func(t *testing.T) {
		g := o.NewWithT(t)

		cm := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-config",
				Namespace: "test-ns",
				Labels: map[string]string{
					annotations.Config: "true",
				},
			},
			Data: map[string]string{
				constants.ConfigFilename: "",
			},
		}

		fakeKube := k8s.NewFakeKube(cm)
		mgr := NewConfigMapManager(fakeKube, "test-app")

		cfg, err := mgr.GetConfig(ctx)
		g.Expect(err).To(o.HaveOccurred())
		g.Expect(err).To(o.MatchError(o.ContainSubstring("invalid configmap found in the cluster")))
		g.Expect(cfg).To(o.BeNil())
	})

	t.Run("error_no_configmap", func(t *testing.T) {
		g := o.NewWithT(t)

		fakeKube := k8s.NewFakeKube()
		mgr := NewConfigMapManager(fakeKube, "test-app")

		cfg, err := mgr.GetConfig(ctx)
		g.Expect(err).To(o.HaveOccurred())
		g.Expect(err).To(o.MatchError(o.ContainSubstring("cluster configmap not found")))
		g.Expect(cfg).To(o.BeNil())
	})
}

// TestCreate tests creating a ConfigMap in the cluster
func TestCreate(t *testing.T) {
	g := o.NewWithT(t)
	ctx := context.Background()

	validConfig := `tssc:
  namespace: create-ns
  settings:
    crc: false
  products: []
`

	cfg, err := NewConfigFromBytes([]byte(validConfig), "create-ns")
	g.Expect(err).To(o.Succeed())

	fakeKube := k8s.NewFakeKube()
	mgr := NewConfigMapManager(fakeKube, "test-app")

	err = mgr.Create(ctx, cfg)
	g.Expect(err).To(o.Succeed())

	// Note: FakeKube doesn't persist changes, so we just verify Create doesn't error
	// The configMapForConfig test verifies the structure of the created ConfigMap
}

// TestUpdate tests updating a ConfigMap in the cluster
func TestUpdate(t *testing.T) {
	g := o.NewWithT(t)
	ctx := context.Background()

	originalConfig := `tssc:
  namespace: update-ns
  settings:
    crc: false
  products: []
`

	// Create initial ConfigMap
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-app-config",
			Namespace: "update-ns",
			Labels: map[string]string{
				annotations.Config: "true",
			},
		},
		Data: map[string]string{
			constants.ConfigFilename: originalConfig,
		},
	}

	fakeKube := k8s.NewFakeKube(cm)
	mgr := NewConfigMapManager(fakeKube, "test-app")

	// Update with new config
	updatedConfigStr := `tssc:
  namespace: update-ns
  settings:
    crc: true
  products: []
`
	cfg, err := NewConfigFromBytes([]byte(updatedConfigStr), "update-ns")
	g.Expect(err).To(o.Succeed())

	err = mgr.Update(ctx, cfg)
	g.Expect(err).To(o.Succeed())

	// Note: FakeKube doesn't persist state changes, so we just verify Update doesn't error
}

// TestDelete tests deleting a ConfigMap from the cluster
func TestDelete(t *testing.T) {
	g := o.NewWithT(t)
	ctx := context.Background()

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-config",
			Namespace: "delete-ns",
			Labels: map[string]string{
				annotations.Config: "true",
			},
		},
		Data: map[string]string{
			constants.ConfigFilename: "test: config",
		},
	}

	fakeKube := k8s.NewFakeKube(cm)
	mgr := NewConfigMapManager(fakeKube, "test-app")

	// Verify ConfigMap exists
	_, err := mgr.GetConfigMap(ctx)
	g.Expect(err).To(o.Succeed())

	// Delete it
	err = mgr.Delete(ctx)
	g.Expect(err).To(o.Succeed())

	// Note: FakeKube doesn't persist state changes, so we just verify Delete doesn't error
}

// TestConfigMapForConfig tests the internal configMapForConfig method
func TestConfigMapForConfig(t *testing.T) {
	g := o.NewWithT(t)

	validConfig := `tssc:
  namespace: test-ns
  settings:
    crc: false
  products: []
`

	cfg, err := NewConfigFromBytes([]byte(validConfig), "test-ns")
	g.Expect(err).To(o.Succeed())

	fakeKube := k8s.NewFakeKube()
	mgr := NewConfigMapManager(fakeKube, "my-app")

	cm := mgr.configMapForConfig(cfg)

	g.Expect(cm).ToNot(o.BeNil())
	g.Expect(cm.Name).To(o.Equal("my-app-config"))
	g.Expect(cm.Namespace).To(o.Equal("test-ns"))
	g.Expect(cm.Labels[annotations.Config]).To(o.Equal("true"))
	g.Expect(cm.Data[constants.ConfigFilename]).To(o.ContainSubstring("tssc:"))
	g.Expect(cm.Data[constants.ConfigFilename]).To(o.Equal(cfg.String()))
}

// TestName tests the Name method
func TestName(t *testing.T) {
	g := o.NewWithT(t)

	fakeKube := k8s.NewFakeKube()

	t.Run("simple_name", func(t *testing.T) {
		mgr := NewConfigMapManager(fakeKube, "app")
		g.Expect(mgr.Name()).To(o.Equal("app-config"))
	})

	t.Run("complex_name", func(t *testing.T) {
		mgr := NewConfigMapManager(fakeKube, "my-complex-app-name")
		g.Expect(mgr.Name()).To(o.Equal("my-complex-app-name-config"))
	})
}

// TestSelector tests the Selector constant
func TestSelector(t *testing.T) {
	g := o.NewWithT(t)

	expectedSelector := annotations.Config + "=true"
	g.Expect(Selector).To(o.Equal(expectedSelector))
	g.Expect(Selector).To(o.ContainSubstring("helmet.redhat-appstudio.github.com/config=true"))
}

// TestConfigMapManagerErrors tests error conditions
func TestConfigMapManagerErrors(t *testing.T) {
	g := o.NewWithT(t)

	t.Run("errors_are_defined", func(t *testing.T) {
		g.Expect(ErrConfigMapNotFound).ToNot(o.BeNil())
		g.Expect(ErrConfigMapNotFound.Error()).To(o.ContainSubstring("cluster configmap not found"))

		g.Expect(ErrMultipleConfigMapFound).ToNot(o.BeNil())
		g.Expect(ErrMultipleConfigMapFound.Error()).To(o.ContainSubstring("multiple cluster configmaps found"))

		g.Expect(ErrIncompleteConfigMap).ToNot(o.BeNil())
		g.Expect(ErrIncompleteConfigMap.Error()).To(o.ContainSubstring("invalid configmap found in the cluster"))
	})
}
