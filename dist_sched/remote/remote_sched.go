package remote

import (
	"context"
	"fmt"
	"os"


	log "github.com/sirupsen/logrus"

	pb "github.com/LucaChot/basic_sched/dist_sched/message"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type RemoteScheduler struct {
    hostname string
    onNode *v1.Node
    clientset   *kubernetes.Clientset
    ctlPlStub  pb.PodPlacementClient
}

func (rmt *RemoteScheduler) SetClientset() {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.WithFields(log.Fields{
			"ERROR": err,
		}).Error("CONFIG ERROR")
	}

	clientset, err := kubernetes.NewForConfig(config)
	rmt.clientset = clientset
}

func (rmt *RemoteScheduler) SetHostname() {
	hostname, err := os.Hostname()
	if err != nil {
		log.WithFields(log.Fields{
			"ERROR": err,
		}).Error("HOSTNAME ERROR")
	}
    rmt.hostname = hostname;
}

func (rmt *RemoteScheduler) SetOnNode() {
    pods, err := rmt.clientset.CoreV1().Pods("basic-sched").List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", rmt.hostname),
	})
	if err != nil {
		log.WithFields(log.Fields{
			"ERROR": err,
		}).Error("LOCATING POD ERROR")
	}

	n, err := rmt.clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", "kubernetes.io/hostname", pods.Items[0].Spec.NodeName),
	})
	if err != nil {
		log.WithFields(log.Fields{
			"ERROR": err,
		}).Error("LOCATING NODE ERROR")
	}

    rmt.onNode = &n.Items[0]
}

/* Creates a new CentralScheduler */
func New() *RemoteScheduler {

    /* Initialise scheduler values */
    rmt := &RemoteScheduler{}

    /* Set the remote scheduler variables */
    rmt.SetClientset()
    rmt.SetHostname()
    rmt.SetOnNode()

    rmt.AsClient()

	return rmt
}

/* Core Scheduling loop */
func (rmt *RemoteScheduler) Schedule() {

    /* Creates a watch interface for all pods that use this scheduler */
	watch, _ := rmt.clientset.CoreV1().Pods("").Watch(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.schedulerName=%s,spec.nodeName=", "central-sched"),
	})

    /* Loops over all new pod events we detect */
	for event := range watch.ResultChan() {
        /* Ignore events where pods have been added */
		if event.Type != "ADDED" {
			continue
		}

		p := event.Object.(*v1.Pod)
		log.WithFields(log.Fields{
			"namespace": p.Namespace,
			"pod":       p.Name,
		}).Debug("BEGIN POD REQUEST")

        rmt.RequestPod(p)
	}
}

