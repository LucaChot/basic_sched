package rand_sched

import (
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const schedulerName = "basic-sched"

/* Returns a random node from nodes list */
func findNodes(nodes *v1.NodeList) *v1.Node {
	return &nodes.Items[rand.Intn(len(nodes.Items))]
}

func Schedule() {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

    /* Creates a watch interface for all pods that use this scheduler */
	watch, _ := clientset.CoreV1().Pods("").Watch(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.schedulerName=%s,spec.nodeName=", schedulerName),
	})

    /* Queries all the nodes in the cluster */
	nodes, _ := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})

    /* Loops over all new pod events we detect */
	for event := range watch.ResultChan() {
        /* Ignore events where pods have been added */
		if event.Type != "ADDED" {
			continue
		}

        start := time.Now().UTC()

		p := event.Object.(*v1.Pod)
		log.WithFields(log.Fields{
			"namespace": p.Namespace,
			"pod":       p.Name,
		}).Debug("found a pod to schedule")

		randomNode := findNodes(nodes)

        /* Create a bind request for the pod and the selected node */
        /* TODO: Check if we are using the latest API version for binds */
		clientset.CoreV1().Pods(p.Namespace).Bind(context.TODO(), &v1.Binding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      p.Name,
				Namespace: p.Namespace,
			},
			Target: v1.ObjectReference{
				APIVersion: "v1",
				Kind:       "Node",
				Name:       randomNode.Name,
			},
		}, metav1.CreateOptions{})

        /* Creates a new event alerting the binding of the pod */
		end := time.Now().UTC()
        nanosecondsSpent := end.Sub(start).Nanoseconds()
        _, err := clientset.CoreV1().Events(p.Namespace).Create(context.TODO(), &v1.Event{
            Action:         "Binding",
            Message:        fmt.Sprintf("Successfully assigned %s/%s to %s", p.Namespace, p.Name, randomNode.Name),
			Reason:         "Scheduled",
            EventTime:      metav1.NewMicroTime(end),
			Type:           "Normal",
            ReportingController: schedulerName,
            ReportingInstance: fmt.Sprintf("%s-dev-k8s-lc869-00", schedulerName),
			InvolvedObject: v1.ObjectReference{
				Kind:      "Pod",
				Name:      p.Name,
				Namespace: p.Namespace,
				UID:       p.UID,
			},
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: p.Name + "-",
                Annotations: map[string]string{
                    "scheduler/nanoseconds": fmt.Sprintf("%d", nanosecondsSpent),
                },
			},
		}, metav1.CreateOptions{})

		log.WithFields(log.Fields{
			"err": err,
		}).Debug("Created event")
	}
}

