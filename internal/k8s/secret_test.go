package k8s

import (
	"context"
	"testing"

	o "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// TestGetSecret tests the GetSecret function
func TestGetSecret(t *testing.T) {
	g := o.NewWithT(t)

	t.Run("get_existing_secret", func(t *testing.T) {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-secret",
				Namespace: "test-namespace",
			},
			Data: map[string][]byte{
				"key": []byte("value"),
			},
		}

		kube := NewFakeKube(secret)
		ctx := context.Background()
		namespacedName := types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "test-secret",
		}

		result, err := GetSecret(ctx, kube, namespacedName)
		g.Expect(err).To(o.Succeed())
		g.Expect(result).ToNot(o.BeNil())
		g.Expect(result.Name).To(o.Equal("test-secret"))
		g.Expect(result.Namespace).To(o.Equal("test-namespace"))
		g.Expect(result.Data["key"]).To(o.Equal([]byte("value")))
	})

	t.Run("get_nonexistent_secret", func(t *testing.T) {
		kube := NewFakeKube()
		ctx := context.Background()
		namespacedName := types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "nonexistent-secret",
		}

		_, err := GetSecret(ctx, kube, namespacedName)
		g.Expect(err).To(o.HaveOccurred())
		// Fake clientset returns an error for nonexistent resources
	})

	t.Run("get_secret_with_multiple_data_keys", func(t *testing.T) {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "multi-key-secret",
				Namespace: "test-namespace",
			},
			Data: map[string][]byte{
				"key1": []byte("value1"),
				"key2": []byte("value2"),
				"key3": []byte("value3"),
			},
		}

		kube := NewFakeKube(secret)
		ctx := context.Background()
		namespacedName := types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "multi-key-secret",
		}

		result, err := GetSecret(ctx, kube, namespacedName)
		g.Expect(err).To(o.Succeed())
		g.Expect(result).ToNot(o.BeNil())
		g.Expect(result.Data).To(o.HaveLen(3))
		g.Expect(result.Data["key1"]).To(o.Equal([]byte("value1")))
		g.Expect(result.Data["key2"]).To(o.Equal([]byte("value2")))
		g.Expect(result.Data["key3"]).To(o.Equal([]byte("value3")))
	})
}

// TestSecretExists tests the SecretExists function
func TestSecretExists(t *testing.T) {
	g := o.NewWithT(t)

	t.Run("secret_exists", func(t *testing.T) {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "existing-secret",
				Namespace: "test-namespace",
			},
		}

		kube := NewFakeKube(secret)
		ctx := context.Background()
		namespacedName := types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "existing-secret",
		}

		exists, err := SecretExists(ctx, kube, namespacedName)
		g.Expect(err).To(o.Succeed())
		g.Expect(exists).To(o.BeTrue())
	})

	t.Run("secret_does_not_exist", func(t *testing.T) {
		kube := NewFakeKube()
		ctx := context.Background()
		namespacedName := types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "nonexistent-secret",
		}

		exists, err := SecretExists(ctx, kube, namespacedName)
		g.Expect(err).To(o.Succeed())
		g.Expect(exists).To(o.BeFalse())
	})

	t.Run("check_secret_in_different_namespace", func(t *testing.T) {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-secret",
				Namespace: "namespace-1",
			},
		}

		kube := NewFakeKube(secret)
		ctx := context.Background()

		// Check in the correct namespace
		namespacedName1 := types.NamespacedName{
			Namespace: "namespace-1",
			Name:      "my-secret",
		}
		exists, err := SecretExists(ctx, kube, namespacedName1)
		g.Expect(err).To(o.Succeed())
		g.Expect(exists).To(o.BeTrue())

		// Check in a different namespace
		namespacedName2 := types.NamespacedName{
			Namespace: "namespace-2",
			Name:      "my-secret",
		}
		exists, err = SecretExists(ctx, kube, namespacedName2)
		g.Expect(err).To(o.Succeed())
		g.Expect(exists).To(o.BeFalse())
	})
}

// TestDeleteSecret tests the DeleteSecret function
func TestDeleteSecret(t *testing.T) {
	g := o.NewWithT(t)

	t.Run("delete_existing_secret", func(t *testing.T) {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deletable-secret",
				Namespace: "test-namespace",
			},
		}

		kube := NewFakeKube(secret)
		ctx := context.Background()
		namespacedName := types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "deletable-secret",
		}

		// Verify secret exists
		exists, err := SecretExists(ctx, kube, namespacedName)
		g.Expect(err).To(o.Succeed())
		g.Expect(exists).To(o.BeTrue())

		// Delete the secret
		err = DeleteSecret(ctx, kube, namespacedName)
		g.Expect(err).To(o.Succeed())
	})

	t.Run("delete_nonexistent_secret", func(t *testing.T) {
		kube := NewFakeKube()
		ctx := context.Background()
		namespacedName := types.NamespacedName{
			Namespace: "test-namespace",
			Name:      "nonexistent-secret",
		}

		// Deleting a nonexistent secret should return an error
		err := DeleteSecret(ctx, kube, namespacedName)
		g.Expect(err).To(o.HaveOccurred())
	})

	t.Run("delete_secret_from_specific_namespace", func(t *testing.T) {
		secret1 := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "shared-name",
				Namespace: "namespace-1",
			},
		}
		secret2 := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "shared-name",
				Namespace: "namespace-2",
			},
		}

		kube := NewFakeKube(secret1, secret2)
		ctx := context.Background()

		// Verify both secrets exist initially
		namespacedName1 := types.NamespacedName{
			Namespace: "namespace-1",
			Name:      "shared-name",
		}
		exists, err := SecretExists(ctx, kube, namespacedName1)
		g.Expect(err).To(o.Succeed())
		g.Expect(exists).To(o.BeTrue())

		namespacedName2 := types.NamespacedName{
			Namespace: "namespace-2",
			Name:      "shared-name",
		}
		exists, err = SecretExists(ctx, kube, namespacedName2)
		g.Expect(err).To(o.Succeed())
		g.Expect(exists).To(o.BeTrue())

		// Delete from namespace-1 should succeed
		err = DeleteSecret(ctx, kube, namespacedName1)
		g.Expect(err).To(o.Succeed())
	})
}
