package k8s

import (
	"context"
	"fmt"
	"log"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/watch"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/krapie/l7/internal/backend/register"
	"github.com/krapie/l7/internal/backend/registry"
)

type Register struct {
	client          *clientset.Clientset
	ServiceRegistry *registry.BackendRegistry
	EventChannel    chan register.BackendEvent

	TargetFilter string
}

func NewRegister() (*Register, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	client, err := clientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Register{
		client: client,

		EventChannel: make(chan register.BackendEvent),
	}, nil
}

func (r *Register) GetEventChannel() chan register.BackendEvent {
	return r.EventChannel
}

func (r *Register) SetTargetFilter(targetFilter string) {
	r.TargetFilter = targetFilter
}

func (r *Register) SetRegistry(registry *registry.BackendRegistry) {
	r.ServiceRegistry = registry
}

func (r *Register) Initialize() error {
	if r.ServiceRegistry == nil {
		return register.ErrRegistryNotSet
	}

	// TODO(krapie): we don't need to initialize the registry here, because watcher will do it
	//pods, err := r.client.CoreV1().Pods("yorkie").List(context.TODO(), metav1.ListOptions{
	//	// TODO(krapie): we temporary hardcode the label selector here
	//	LabelSelector: labels.Set(map[string]string{"app.kubernetes.io/name": "yorkie"}).AsSelector().String(),
	//})
	//if err != nil {
	//	return err
	//}
	//
	//for _, pod := range pods.Items {
	//	err = r.ServiceRegistry.AddBackend(
	//		pod.Name,
	//		fmt.Sprintf("%s://%s:%d", register.SCHEME, pod.Status.PodIP, pod.Spec.Containers[0].Ports[0].ContainerPort),
	//	)
	//	if err != nil {
	//		return err
	//	}
	//
	//	r.EventChannel <- register.BackendEvent{
	//		EventType: register.BackendAddedEvent,
	//		Actor:     pod.Name,
	//	}
	//
	//	log.Printf("Pod Hostname: %s, IP: %s:%s:%d\n", pod.ObjectMeta.Name, pod.Status.PodIP, pod.Spec.Containers[0].Ports[0].ContainerPort)
	//}

	return nil
}

func (r *Register) Observe() {
	go r.observe()
}

func (r *Register) observe() {
	podWatcher, err := r.client.CoreV1().Pods(r.TargetFilter).Watch(context.TODO(), metav1.ListOptions{
		LabelSelector: labels.Set(map[string]string{"app.kubernetes.io/instance": r.TargetFilter}).AsSelector().String(),
	})
	if err != nil {
		log.Println("[Register] Error:", err)
	}

	for event := range podWatcher.ResultChan() {
		switch event.Type {
		case watch.Added, watch.Modified:
			pod, ok := event.Object.(*corev1.Pod)
			if !ok {
				log.Println("[Register] Error: failed to cast to *corev1.Pod")
				continue
			}

			if pod.Status.PodIP == "" {
				// log.Println("[Register] Error: pod IP is empty")
				continue
			}

			err = r.ServiceRegistry.AddBackend(
				pod.Name,
				fmt.Sprintf("%s://%s:%d", register.SCHEME, pod.Status.PodIP, pod.Spec.Containers[0].Ports[0].ContainerPort),
			)
			if err != nil {
				// log.Printf("[Register] Error adding backend: %s", err)
				continue
			}
			r.EventChannel <- register.BackendEvent{
				EventType: register.BackendAddedEvent,
				Actor:     pod.Name,
			}
			log.Printf("[Register] Backend Added: Hostname: %s, IP: %s:%d\n", pod.Name, pod.Status.PodIP, pod.Spec.Containers[0].Ports[0].ContainerPort)
		case watch.Deleted:
			pod, ok := event.Object.(*corev1.Pod)
			if !ok {
				log.Println("[Register] Error: failed to cast to *corev1.Pod")
				continue
			}
			r.ServiceRegistry.RemoveBackendByID(pod.Name)
			r.EventChannel <- register.BackendEvent{
				EventType: register.BackendRemovedEvent,
				Actor:     pod.Name,
			}
			log.Printf("[Register] Backend Removed: Hostname: %s, IP: %s:%d\n", pod.Name, pod.Status.PodIP, pod.Spec.Containers[0].Ports[0].ContainerPort)
		case watch.Error:
			log.Println("[Register] Error:", err)
		}
	}
}
