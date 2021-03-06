package gube

import (
	"fmt"
	"sync"

	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

type Shooted interface {
	AsShoot() (Shoot, error)
	GetShootName() *ShootName
}

type Cluster interface {
	GetClientConfig() (*restclient.Config, error)
	GetClientset() (*kubernetes.Clientset, error)

	GetNodeCount() (int, error)
	GetNodes() (map[string]corev1.Node, error)
	GetPodCount() (int, error)
	GetPods(namespace string) (map[string]corev1.Pod, error)
	GetSecretByRef(secretref corev1.SecretReference) (*corev1.Secret, error)
	GetIngress(name, ns string) (*extv1beta1.Ingress, error)
	GetConfigMap(name, ns string) (*corev1.ConfigMap, error)
	GetConfigMapEntries(name, ns string) (map[string]string, error)
	KubeconfigProvider
	Shooted

	GetClusterKey() string
}

type cluster struct {
	access    KubeconfigProvider
	config    *restclient.Config
	clientset *kubernetes.Clientset
	key       string
	lock      sync.Mutex
}

func NewCluster(key string, cfg KubeconfigProvider) Cluster {
	return (&cluster{}).new(key, cfg)
}

func (this *cluster) new(key string, cfg KubeconfigProvider) *cluster {
	this.access = cfg
	this.key = key
	return this
}

func (this *cluster) GetClusterKey() string {
	return this.key
}

func (this *cluster) AsShoot() (Shoot, error) {
	return nil, fmt.Errorf("%s not shooted", this.GetClusterKey())
}
func (this *cluster) GetShootName() *ShootName {
	return nil
}

func (this *cluster) GetKubeconfig() ([]byte, error) {
	return this.access.GetKubeconfig()
}

func (this *cluster) GetClientConfig() (*restclient.Config, error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	if this.config == nil {
		bytes, err := this.GetKubeconfig()
		if err != nil {
			return nil, err
		}
		this.config, err = NewConfigFromBytes(bytes)
		if err != nil {
			return nil, err
		}
	}
	return this.config, nil
}

func (this *cluster) GetClientset() (*kubernetes.Clientset, error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	if this.clientset == nil {
		bytes, err := this.GetKubeconfig()
		if err != nil {
			return nil, err
		}
		this.clientset, err = NewClientFromBytes(bytes)
		if err != nil {
			return nil, err
		}
	}
	return this.clientset, nil
}

func (this *cluster) GetNodeCount() (int, error) {
	cs, err := this.GetClientset()
	if err != nil {
		return 0, err
	}
	list, err := cs.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to get node count for %s: %s", this.GetClusterKey(), err)
	}
	return len(list.Items), nil
}

func (this *cluster) GetNodes() (map[string]corev1.Node, error) {
	cs, err := this.GetClientset()
	if err != nil {
		return nil, err
	}
	list, err := cs.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes for %s: %s", this.GetClusterKey(), err)
	}
	nodes := map[string]corev1.Node{}
	for _, n := range list.Items {
		nodes[n.GetName()] = n
	}
	return nodes, nil
}

func (this *cluster) GetPodCount() (int, error) {
	cs, err := this.GetClientset()
	if err != nil {
		return 0, err
	}
	list, err := cs.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		return 0, fmt.Errorf("failed to get pod count for %s: %s", this.GetClusterKey(), err)
	}
	return len(list.Items), nil
}

func (this *cluster) GetPods(namespace string) (map[string]corev1.Pod, error) {
	cs, err := this.GetClientset()
	if err != nil {
		return nil, err
	}
	list, err := cs.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pods for %s: %s", this.GetClusterKey(), err)
	}
	pods := map[string]corev1.Pod{}
	for _, n := range list.Items {
		pods[n.GetName()] = n
	}
	return pods, nil
}

func (this *cluster) GetSecretByRef(secretref corev1.SecretReference) (*corev1.Secret, error) {
	kubeset, err := this.GetClientset()
	if err != nil {
		return nil, fmt.Errorf("failed to get secret for %s: %s", this.GetClusterKey(), err)
	}
	secret, err := kubeset.CoreV1().Secrets(secretref.Namespace).Get(secretref.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret %s for namespace %s for %s: %s",
			secretref.Name, secretref.Namespace, this.GetClusterKey(), err)
	}
	return secret, nil
}

func (this *cluster) GetConfigMap(name, ns string) (*corev1.ConfigMap, error) {
	kubeset, err := this.GetClientset()
	if err != nil {
		return nil, fmt.Errorf("failed to get config map for %s: %s", this.GetClusterKey(), err)
	}
	config, err := kubeset.CoreV1().ConfigMaps(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get config map %s for namespace %s for %s: %s",
			name, ns, this.GetClusterKey(), err)
	}
	return config, nil
}

func (this *cluster) GetIngress(name, ns string) (*extv1beta1.Ingress, error) {
	kubeset, err := this.GetClientset()
	if err != nil {
		return nil, fmt.Errorf("failed to get ingress for %s: %s", this.GetClusterKey(), err)
	}
	ingress, err := kubeset.ExtensionsV1beta1().Ingresses(ns).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get ingress %s for namespace %s for %s: %s",
			name, ns, this.GetClusterKey(), err)
	}
	return ingress, nil
}

func (this *cluster) GetConfigMapEntries(name, ns string) (map[string]string, error) {
	config, err := this.GetConfigMap(name, ns)
	if err != nil {
		return nil, err
	}
	return config.Data, nil
}
