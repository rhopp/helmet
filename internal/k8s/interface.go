package k8s

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	batchv1client "k8s.io/client-go/kubernetes/typed/batch/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	rbacv1client "k8s.io/client-go/kubernetes/typed/rbac/v1"
)

type Interface interface {
	BatchV1ClientSet(string) (batchv1client.BatchV1Interface, error)
	ClientSet(string) (kubernetes.Interface, error)
	Connected() error
	CoreV1ClientSet(string) (corev1client.CoreV1Interface, error)
	DiscoveryClient(string) (discovery.DiscoveryInterface, error)
	DynamicClient(string) (dynamic.Interface, error)
	GetDynamicClientForObjectRef(*corev1.ObjectReference) (dynamic.ResourceInterface, error)
	RBACV1ClientSet(string) (rbacv1client.RbacV1Interface, error)
	RESTClientGetter(string) genericclioptions.RESTClientGetter
}
