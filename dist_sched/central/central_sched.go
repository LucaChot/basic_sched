package central

import (
	"context"
	"fmt"
	"time"
    "sync"

	log "github.com/sirupsen/logrus"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"
	pb "github.com/LucaChot/basic_sched/dist_sched/message"
)

type CentralScheduler struct {
    mu          sync.Mutex
    Name        string
    clientset   *kubernetes.Clientset
    Bins        map[string]string
    pb.UnimplementedPodPlacementServer
}

/* Creates a new CentralScheduler */
func New() *CentralScheduler {

    /* Initialise scheduler values */
	ctl := &CentralScheduler{
		Name: "central-sched",
        Bins: make(map[string]string),
    }

	config, err := rest.InClusterConfig()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("error getting config")
	}
	config.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(80, 100)

	clientset, err := kubernetes.NewForConfig(config)
	ctl.clientset = clientset

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("error getting config")
	}

    go ctl.ctlStartPlacementServer()

	return ctl
}


/* Returns the node of the first remote scheduler to respond to new pod */
func (ctl *CentralScheduler) findNode(podName string) string {
    /* Busy wait until a scheduler has responded */
    for {
        if _, ok := ctl.Bins[podName]; ok {
            break
        }
    }
    /* Fetch the node to place the pod */
    return ctl.Bins[podName]
}

/* Core Scheduling loop */
func (ctl *CentralScheduler) Schedule() {

    /* Creates a watch interface for all pods that use this scheduler */
	watch, _ := ctl.clientset.CoreV1().Pods("").Watch(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.schedulerName=%s,spec.nodeName=", ctl.Name),
	})

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
		}).Debug("BEGIN POD SCHEDULE")


        /* Find a node to place the pod */
        node := ctl.findNode(p.Name)
        ctl.placePodToNode(p, node)

        /* Collect information for event */
		end := time.Now().UTC()
        nanosecondsSpent := end.Sub(start).Nanoseconds()
        annotations := map[string]string{
            "scheduler/nanoseconds": fmt.Sprintf("%d", nanosecondsSpent),
        }

        /* Creates a new event alerting the binding of the pod */
        err := ctl.createSchedEvent(p, node, end, annotations)
        if err != nil {
            log.WithFields(log.Fields{
                "err": err,
            }).Debug("FAILED TO CREATE EVENT")
        }
	}
}

