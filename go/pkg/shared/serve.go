package shared

// TODO(andreas): document this for R API etc

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/replicate/replicate/go/pkg/console"
	"github.com/replicate/replicate/go/pkg/errors"
	"github.com/replicate/replicate/go/pkg/project"
	"github.com/replicate/replicate/go/pkg/servicepb"
)

type projectGetter func() (proj *project.Project, err error)

type server struct {
	servicepb.UnimplementedDaemonServer

	workChan                 chan func() error
	projectGetter            projectGetter
	project                  *project.Project
	heartbeatsByExperimentID map[string]*HeartbeatProcess
}

func (s *server) CreateExperiment(ctx context.Context, req *servicepb.CreateExperimentRequest) (*servicepb.CreateExperimentReply, error) {
	pbReqExp := req.GetExperiment()
	args := project.CreateExperimentArgs{
		Path:           pbReqExp.GetPath(),
		Command:        pbReqExp.GetCommand(),
		Params:         valueMapFromPb(pbReqExp.GetParams()),
		PythonPackages: pbReqExp.GetPythonPackages(),
	}
	proj, err := s.getProject()
	if err != nil {
		return nil, handleError(err)
	}
	exp, err := proj.CreateExperiment(args, true, s.workChan)
	if err != nil {
		return nil, handleError(err)
	}
	if !req.DisableHeartbeat {
		s.heartbeatsByExperimentID[exp.ID] = StartHeartbeat(s.project, exp.ID)
	}

	pbRetExp := experimentToPb(exp)
	return &servicepb.CreateExperimentReply{Experiment: pbRetExp}, nil
}

func (s *server) CreateCheckpoint(ctx context.Context, req *servicepb.CreateCheckpointRequest) (*servicepb.CreateCheckpointReply, error) {
	pbReqChk := req.GetCheckpoint()
	args := project.CreateCheckpointArgs{
		Path:          pbReqChk.GetPath(),
		Metrics:       valueMapFromPb(pbReqChk.GetMetrics()),
		PrimaryMetric: primaryMetricFromPb(pbReqChk.PrimaryMetric),
		Step:          pbReqChk.GetStep(),
	}
	proj, err := s.getProject()
	if err != nil {
		return nil, handleError(err)
	}
	chk, err := proj.CreateCheckpoint(args, true, s.workChan, req.Quiet)
	if err != nil {
		return nil, handleError(err)
	}

	pbRetChk := checkpointToPb(chk)
	return &servicepb.CreateCheckpointReply{Checkpoint: pbRetChk}, nil
}

func (s *server) SaveExperiment(ctx context.Context, req *servicepb.SaveExperimentRequest) (*servicepb.SaveExperimentReply, error) {
	expPb := req.GetExperiment()
	exp := experimentFromPb(expPb)
	proj, err := s.getProject()
	if err != nil {
		return nil, handleError(err)
	}
	exp, err = proj.SaveExperiment(exp)
	if err != nil {
		return nil, handleError(err)
	}
	return &servicepb.SaveExperimentReply{Experiment: experimentToPb(exp)}, nil
}

func (s *server) StopExperiment(ctx context.Context, req *servicepb.StopExperimentRequest) (*servicepb.StopExperimentReply, error) {
	if _, ok := s.heartbeatsByExperimentID[req.ExperimentID]; ok {
		s.heartbeatsByExperimentID[req.ExperimentID].Kill()
		delete(s.heartbeatsByExperimentID, req.ExperimentID)
	}
	proj, err := s.getProject()
	if err != nil {
		return nil, handleError(err)
	}
	if err := proj.StopExperiment(req.ExperimentID); err != nil {
		return nil, handleError(err)
	}
	return &servicepb.StopExperimentReply{}, nil
}

func (s *server) GetExperiment(ctx context.Context, req *servicepb.GetExperimentRequest) (*servicepb.GetExperimentReply, error) {
	proj, err := s.getProject()
	if err != nil {
		return nil, handleError(err)
	}
	exp, err := proj.ExperimentFromPrefix(req.ExperimentIDPrefix)
	if err != nil {
		return nil, handleError(err)
	}
	expPb := experimentToPb(exp)
	return &servicepb.GetExperimentReply{Experiment: expPb}, nil
}

func (s *server) ListExperiments(ctx context.Context, req *servicepb.ListExperimentsRequest) (*servicepb.ListExperimentsReply, error) {
	proj, err := s.getProject()
	if err != nil {
		return nil, handleError(err)
	}
	experiments, err := proj.Experiments()
	if err != nil {
		return nil, handleError(err)
	}
	experimentsPb := experimentsToPb(experiments)
	return &servicepb.ListExperimentsReply{Experiments: experimentsPb}, nil
}

func (s *server) DeleteExperiment(ctx context.Context, req *servicepb.DeleteExperimentRequest) (*servicepb.DeleteExperimentReply, error) {
	proj, err := s.getProject()
	if err != nil {
		return nil, handleError(err)
	}
	exp, err := proj.ExperimentByID(req.ExperimentID)
	if err != nil {
		return nil, handleError(err)
	}
	if err := s.project.DeleteExperiment(exp); err != nil {
		return nil, handleError(err)
	}
	// This is slow, see https://github.com/replicate/replicate/issues/333
	for _, checkpoint := range exp.Checkpoints {
		if err := s.project.DeleteCheckpoint(checkpoint); err != nil {
			return nil, handleError(err)
		}
	}

	return &servicepb.DeleteExperimentReply{}, nil
}

func (s *server) CheckoutCheckpoint(ctx context.Context, req *servicepb.CheckoutCheckpointRequest) (*servicepb.CheckoutCheckpointReply, error) {
	proj, err := s.getProject()
	if err != nil {
		return nil, handleError(err)
	}
	chk, exp, err := proj.CheckpointFromPrefix(req.CheckpointIDPrefix)
	if err != nil {
		return nil, handleError(err)
	}

	err = s.project.CheckoutCheckpoint(chk, exp, req.OutputDirectory)
	if err != nil {
		return nil, handleError(err)
	}
	return &servicepb.CheckoutCheckpointReply{}, nil
}

func (s *server) GetExperimentStatus(ctx context.Context, req *servicepb.GetExperimentStatusRequest) (*servicepb.GetExperimentStatusReply, error) {
	proj, err := s.getProject()
	if err != nil {
		return nil, handleError(err)
	}
	isRunning, err := proj.ExperimentIsRunning(req.ExperimentID)
	if err != nil {
		return nil, handleError(err)
	}
	var status servicepb.GetExperimentStatusReply_Status
	if isRunning {
		status = servicepb.GetExperimentStatusReply_RUNNING
	} else {
		status = servicepb.GetExperimentStatusReply_STOPPED
	}
	return &servicepb.GetExperimentStatusReply{Status: status}, nil
}

func (s *server) getProject() (*project.Project, error) {
	// we get the project lazily so that we can return a protobuf exception to the client
	// as part of a request flow

	if s.project != nil {
		return s.project, nil
	}

	proj, err := s.projectGetter()
	if err != nil {
		return nil, err
	}
	s.project = proj
	return proj, nil
}

func Serve(projGetter projectGetter, socketPath string) error {
	console.Info("Starting daemon")

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("Failed to open UNIX socket on %s: %w", socketPath, err)
	}

	grpcServer := grpc.NewServer()
	s := &server{
		// block if there already are two items on the queue, in case uploading is a bottleneck
		// TODO(andreas): warn the user if the queue is full, so they know that they should
		// upload at a lesser interval
		workChan:                 make(chan func() error, 2),
		projectGetter:            projGetter,
		heartbeatsByExperimentID: make(map[string]*HeartbeatProcess),
	}
	servicepb.RegisterDaemonServer(grpcServer, s)

	// when the process exits, make sure any pending
	// uploads are completed
	exitChan := make(chan struct{})
	completedChan := make(chan struct{})

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func() {
		<-sigc
		console.Info("Exiting...")
		exitChan <- struct{}{}
		for _, hb := range s.heartbeatsByExperimentID {
			hb.Kill()
		}
		grpcServer.GracefulStop()
		<-completedChan
	}()

	go func() {
		for {
			select {
			case work := <-s.workChan:
				// TODO(andreas): fail hard if an error occurs in CreateExperiment
				if err := work(); err != nil {
					console.Error("%v", err)
					// TODO(andreas): poll status endpoint, put errors in chan of messages to return. also include progress in these messages
				}
			case <-exitChan:
				completedChan <- struct{}{}
				return
			}
		}
	}()

	if err := grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("Failed to start server: %w", err)
	}

	return nil
}

func handleError(err error) error {
	reason := errors.Code(err)
	if reason != "" {
		st := status.New(codes.Internal, err.Error())
		details := &errdetails.ErrorInfo{Reason: reason}
		st, err := st.WithDetails(details)
		if err != nil {
			return err
		}
		return st.Err()
	}
	return status.Error(codes.Unknown, err.Error())
}
