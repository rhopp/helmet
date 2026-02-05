package k8s

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/redhat-appstudio/helmet/internal/annotations"

	o "github.com/onsi/gomega"
	rbacv1 "k8s.io/api/rbac/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestRetry tests the Retry utility function
func TestRetry(t *testing.T) {
	g := o.NewWithT(t)

	t.Run("retry_succeeds_on_first_attempt", func(t *testing.T) {
		attempts := 0
		fn := func() error {
			attempts++
			return nil
		}

		err := Retry(3, 1*time.Millisecond, fn)
		g.Expect(err).To(o.Succeed())
		g.Expect(attempts).To(o.Equal(1))
	})

	t.Run("retry_succeeds_on_second_attempt", func(t *testing.T) {
		attempts := 0
		fn := func() error {
			attempts++
			if attempts < 2 {
				return errors.New("temporary error")
			}
			return nil
		}

		err := Retry(3, 1*time.Millisecond, fn)
		g.Expect(err).To(o.Succeed())
		g.Expect(attempts).To(o.Equal(2))
	})

	t.Run("retry_succeeds_on_last_attempt", func(t *testing.T) {
		attempts := 0
		fn := func() error {
			attempts++
			if attempts < 3 {
				return errors.New("temporary error")
			}
			return nil
		}

		err := Retry(3, 1*time.Millisecond, fn)
		g.Expect(err).To(o.Succeed())
		g.Expect(attempts).To(o.Equal(3))
	})

	t.Run("retry_fails_after_max_attempts", func(t *testing.T) {
		attempts := 0
		expectedErr := errors.New("persistent error")
		fn := func() error {
			attempts++
			return expectedErr
		}

		err := Retry(3, 1*time.Millisecond, fn)
		g.Expect(err).To(o.HaveOccurred())
		g.Expect(err).To(o.Equal(expectedErr))
		g.Expect(attempts).To(o.Equal(3))
	})

	t.Run("retry_with_single_attempt", func(t *testing.T) {
		attempts := 0
		fn := func() error {
			attempts++
			return errors.New("error")
		}

		err := Retry(1, 1*time.Millisecond, fn)
		g.Expect(err).To(o.HaveOccurred())
		g.Expect(attempts).To(o.Equal(1))
	})

	t.Run("retry_with_zero_attempts", func(t *testing.T) {
		attempts := 0
		fn := func() error {
			attempts++
			return errors.New("error")
		}

		// Zero attempts should still try once (attempts - 1 check in loop)
		err := Retry(0, 1*time.Millisecond, fn)
		g.Expect(err).To(o.HaveOccurred())
		g.Expect(attempts).To(o.Equal(1))
	})
}

// TestDeleteClusterRoleBindings tests the DeleteClusterRoleBindings function
func TestDeleteClusterRoleBindings(t *testing.T) {
	g := o.NewWithT(t)

	t.Run("delete_cluster_role_bindings_with_label", func(t *testing.T) {
		// Create test ClusterRoleBindings with the deletion label
		crb1 := &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-crb-1",
				Labels: map[string]string{
					annotations.PostDeploy: "delete",
				},
			},
		}
		crb2 := &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-crb-2",
				Labels: map[string]string{
					annotations.PostDeploy: "delete",
				},
			},
		}

		kube := NewFakeKube(crb1, crb2)
		ctx := context.Background()

		// This should succeed (FakeKube supports DeleteCollection)
		err := DeleteClusterRoleBindings(ctx, kube, "test-namespace")
		g.Expect(err).To(o.Succeed())
	})

	t.Run("delete_cluster_role_bindings_empty_list", func(t *testing.T) {
		kube := NewFakeKube()
		ctx := context.Background()

		err := DeleteClusterRoleBindings(ctx, kube, "test-namespace")
		g.Expect(err).To(o.Succeed())
	})
}

// TestDeleteClusterRoles tests the DeleteClusterRoles function
func TestDeleteClusterRoles(t *testing.T) {
	g := o.NewWithT(t)

	t.Run("delete_cluster_roles_with_label", func(t *testing.T) {
		// Create test ClusterRoles with the deletion label
		cr1 := &rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-cr-1",
				Labels: map[string]string{
					annotations.PostDeploy: "delete",
				},
			},
		}

		kube := NewFakeKube(cr1)
		ctx := context.Background()

		err := DeleteClusterRoles(ctx, kube, "test-namespace")
		g.Expect(err).To(o.Succeed())
	})

	t.Run("delete_cluster_roles_empty_list", func(t *testing.T) {
		kube := NewFakeKube()
		ctx := context.Background()

		err := DeleteClusterRoles(ctx, kube, "test-namespace")
		g.Expect(err).To(o.Succeed())
	})
}

// TestDeleteRoleBindings tests the DeleteRoleBindings function
func TestDeleteRoleBindings(t *testing.T) {
	g := o.NewWithT(t)

	t.Run("delete_role_bindings_with_label", func(t *testing.T) {
		// Create test RoleBindings with the deletion label
		rb1 := &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-rb-1",
				Namespace: "namespace-1",
				Labels: map[string]string{
					annotations.PostDeploy: "delete",
				},
			},
		}
		rb2 := &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-rb-2",
				Namespace: "namespace-2",
				Labels: map[string]string{
					annotations.PostDeploy: "delete",
				},
			},
		}

		kube := NewFakeKube(rb1, rb2)
		ctx := context.Background()

		err := DeleteRoleBindings(ctx, kube, "test-namespace")
		g.Expect(err).To(o.Succeed())
	})

	t.Run("delete_role_bindings_empty_list", func(t *testing.T) {
		kube := NewFakeKube()
		ctx := context.Background()

		err := DeleteRoleBindings(ctx, kube, "test-namespace")
		g.Expect(err).To(o.Succeed())
	})
}

// TestDeleteRoles tests the DeleteRoles function
func TestDeleteRoles(t *testing.T) {
	g := o.NewWithT(t)

	t.Run("delete_roles_with_label", func(t *testing.T) {
		// Create test Roles with the deletion label
		role1 := &rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-role-1",
				Namespace: "namespace-1",
				Labels: map[string]string{
					annotations.PostDeploy: "delete",
				},
			},
		}

		kube := NewFakeKube(role1)
		ctx := context.Background()

		err := DeleteRoles(ctx, kube, "test-namespace")
		g.Expect(err).To(o.Succeed())
	})

	t.Run("delete_roles_empty_list", func(t *testing.T) {
		kube := NewFakeKube()
		ctx := context.Background()

		err := DeleteRoles(ctx, kube, "test-namespace")
		g.Expect(err).To(o.Succeed())
	})
}

// TestDeleteServiceAccounts tests the DeleteServiceAccounts function
func TestDeleteServiceAccounts(t *testing.T) {
	g := o.NewWithT(t)

	t.Run("delete_service_accounts_with_label", func(t *testing.T) {
		// Create test ServiceAccounts with the deletion label
		sa1 := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-sa-1",
				Namespace: "namespace-1",
				Labels: map[string]string{
					annotations.PostDeploy: "delete",
				},
			},
		}

		kube := NewFakeKube(sa1)
		ctx := context.Background()

		err := DeleteServiceAccounts(ctx, kube, "test-namespace")
		g.Expect(err).To(o.Succeed())
	})

	t.Run("delete_service_accounts_empty_list", func(t *testing.T) {
		kube := NewFakeKube()
		ctx := context.Background()

		err := DeleteServiceAccounts(ctx, kube, "test-namespace")
		g.Expect(err).To(o.Succeed())
	})
}

// TestDeleteResources tests the DeleteResources orchestration function
func TestDeleteResources(t *testing.T) {
	g := o.NewWithT(t)

	t.Run("delete_all_resource_types", func(t *testing.T) {
		// Create various resources with the deletion label
		crb := &rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-crb",
				Labels: map[string]string{
					annotations.PostDeploy: "delete",
				},
			},
		}
		cr := &rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-cr",
				Labels: map[string]string{
					annotations.PostDeploy: "delete",
				},
			},
		}
		rb := &rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-rb",
				Namespace: "test-namespace",
				Labels: map[string]string{
					annotations.PostDeploy: "delete",
				},
			},
		}
		role := &rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-role",
				Namespace: "test-namespace",
				Labels: map[string]string{
					annotations.PostDeploy: "delete",
				},
			},
		}
		sa := &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-sa",
				Namespace: "test-namespace",
				Labels: map[string]string{
					annotations.PostDeploy: "delete",
				},
			},
		}

		kube := NewFakeKube(crb, cr, rb, role, sa)
		ctx := context.Background()

		// DeleteResources should call all deletion functions
		err := DeleteResources(ctx, kube, "test-namespace")
		g.Expect(err).To(o.Succeed())
	})

	t.Run("delete_resources_with_no_resources", func(t *testing.T) {
		kube := NewFakeKube()
		ctx := context.Background()

		err := DeleteResources(ctx, kube, "test-namespace")
		g.Expect(err).To(o.Succeed())
	})
}

// TestRetryDeleteResources tests the RetryDeleteResources wrapper function
func TestRetryDeleteResources(t *testing.T) {
	g := o.NewWithT(t)

	t.Run("retry_delete_resources_succeeds", func(t *testing.T) {
		kube := NewFakeKube()
		ctx := context.Background()

		// This should succeed without retries
		err := RetryDeleteResources(ctx, kube, "test-namespace")
		g.Expect(err).To(o.Succeed())
	})
}
