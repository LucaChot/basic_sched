package central

import (
	"context"
	"net"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	pb "github.com/LucaChot/basic_sched/dist_sched/message"
)



func (ctl *CentralScheduler) ctlStartPlacementServer() {
    lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatalf("failed to serve start server")
	}

	s := grpc.NewServer()
    pb.RegisterPodPlacementServer(s, ctl)

	log.WithFields(log.Fields{
		"ADDRESS": lis.Addr(),
	}).Debug("STARTED PLACEMENT SERVER")

	go func() {
		if err := s.Serve(lis); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatalf("failed to serve start server")
		}
	}()
}

func (ctl *CentralScheduler) RequestPod(ctx context.Context, in *pb.PodRequest) (*pb.EmptyReply, error) {
    log.WithFields(log.Fields{
        "POD":      in.Pod,
        "NODE":     in.Node,
    }).Debug("RECEIVED POD REQUEST")
    if _, err := ctl.Bins[in.Pod]; err {
        return &pb.EmptyReply{}, nil
    }
    if ctl.mu.TryLock() {
        defer ctl.mu.Unlock()

        if _, err := ctl.Bins[in.Pod]; err {
            return &pb.EmptyReply{}, nil
        }
        ctl.Bins[in.Pod] = in.Node
    }
    return &pb.EmptyReply{}, nil
}

