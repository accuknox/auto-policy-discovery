package server

import (
	"context"
	"errors"

	"github.com/rs/zerolog"

	analyzer "github.com/accuknox/auto-policy-discovery/src/analyzer"
	cfg "github.com/accuknox/auto-policy-discovery/src/config"
	core "github.com/accuknox/auto-policy-discovery/src/config"
	"github.com/accuknox/auto-policy-discovery/src/feedconsumer"
	logger "github.com/accuknox/auto-policy-discovery/src/logging"
	networker "github.com/accuknox/auto-policy-discovery/src/networkpolicy"
	sysworker "github.com/accuknox/auto-policy-discovery/src/systempolicy"

	"github.com/accuknox/auto-policy-discovery/src/libs"
	apb "github.com/accuknox/auto-policy-discovery/src/protobuf/v1/analyzer"
	fpb "github.com/accuknox/auto-policy-discovery/src/protobuf/v1/consumer"
	opb "github.com/accuknox/auto-policy-discovery/src/protobuf/v1/observability"
	wpb "github.com/accuknox/auto-policy-discovery/src/protobuf/v1/worker"
	"github.com/accuknox/auto-policy-discovery/src/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

const PortNumber = "9089"

var log *zerolog.Logger

func init() {
	log = logger.GetInstance()
}

// ==================== //
// == Worker Service == //
// ==================== //

type workerServer struct {
	wpb.WorkerServer
}

func (s *workerServer) Start(ctx context.Context, in *wpb.WorkerRequest) (*wpb.WorkerResponse, error) {
	log.Info().Msg("Start worker called")

	response := ""

	if in.GetReq() == "dbclear" {
		libs.ClearDBTables(core.CurrentCfg.ConfigDB)
		response += "Cleared DB."
	}

	if in.GetLogfile() != "" {
		core.SetLogFile(in.GetLogfile())
		response += "Log File Set ,"
	}

	if in.GetPolicytype() != "" {
		if in.GetPolicytype() == "network" {
			networker.StartNetworkWorker()
		} else if in.GetPolicytype() == "system" {
			sysworker.StartSystemWorker()
		}
		response += "Starting " + in.GetPolicytype() + " policy discovery"
	}

	return &wpb.WorkerResponse{Res: response}, nil
}

func (s *workerServer) Stop(ctx context.Context, in *wpb.WorkerRequest) (*wpb.WorkerResponse, error) {
	log.Info().Msg("Stop worker called")

	if in.GetPolicytype() == "network" {
		networker.StopNetworkWorker()
	} else if in.GetPolicytype() == "system" {
		sysworker.StopSystemWorker()
	} else {
		return &wpb.WorkerResponse{Res: "No policy type, choose 'network' or 'system', not [" + in.GetPolicytype() + "]"}, nil
	}

	return &wpb.WorkerResponse{Res: "ok stopping " + in.GetPolicytype() + " policy discovery"}, nil
}

func (s *workerServer) GetWorkerStatus(ctx context.Context, in *wpb.WorkerRequest) (*wpb.WorkerResponse, error) {
	log.Info().Msg("Get worker status called")

	status := ""

	if in.GetPolicytype() == "network" {
		status = networker.NetworkWorkerStatus
	} else if in.GetPolicytype() == "system" {
		status = sysworker.SystemWorkerStatus
	} else {
		return &wpb.WorkerResponse{Res: "No policy type, choose 'network' or 'system', not [" + in.GetPolicytype() + "]"}, nil
	}

	return &wpb.WorkerResponse{Res: status}, nil
}

func (s *workerServer) Convert(ctx context.Context, in *wpb.WorkerRequest) (*wpb.WorkerResponse, error) {

	if in.GetPolicytype() == "network" {
		log.Info().Msg("Convert network policy called")
		networker.InitNetPolicyDiscoveryConfiguration()
		return networker.GetNetPolicy(in.Clustername, in.Namespace), nil
	} else if in.GetPolicytype() == "system" {
		log.Info().Msg("Convert system policy called")
		sysworker.InitSysPolicyDiscoveryConfiguration()
		return sysworker.GetSysPolicy(in.Namespace, in.Clustername, in.Labels, in.Fromsource), nil
	} else {
		log.Info().Msg("Convert policy called, but no policy type")
	}

	return &wpb.WorkerResponse{Res: "ok"}, nil
}

// ====================== //
// == Consumer Service == //
// ====================== //

type consumerServer struct {
	fpb.ConsumerServer
}

func (s *consumerServer) Start(ctx context.Context, in *fpb.ConsumerRequest) (*fpb.ConsumerResponse, error) {
	log.Info().Msg("Start consumer called")
	feedconsumer.StartConsumer()
	return &fpb.ConsumerResponse{Res: "ok"}, nil
}

func (s *consumerServer) Stop(ctx context.Context, in *fpb.ConsumerRequest) (*fpb.ConsumerResponse, error) {
	log.Info().Msg("Stop consumer called")
	feedconsumer.StopConsumer()
	return &fpb.ConsumerResponse{Res: "ok"}, nil
}

func (s *consumerServer) GetWorkerStatus(ctx context.Context, in *fpb.ConsumerRequest) (*fpb.ConsumerResponse, error) {
	log.Info().Msg("Get consumer status called")
	return &fpb.ConsumerResponse{Res: feedconsumer.Status}, nil
}

// ====================== //
// == Analyzer Service == //
// ====================== //

type analyzerServer struct {
	apb.AnalyzerServer
}

func (s *analyzerServer) GetNetworkPolicies(ctx context.Context, in *apb.NetworkLogs) (*apb.NetworkPolicies, error) {
	pbNetworkPolicies := apb.NetworkPolicies{}
	pbNetworkPolicies.NwPolicies = analyzer.GetNetworkPolicies(in.GetNwLog())
	return &pbNetworkPolicies, nil
}

func (s *analyzerServer) GetSystemPolicies(ctx context.Context, in *apb.SystemLogs) (*apb.SystemPolicies, error) {
	pbSystemPolicies := apb.SystemPolicies{}
	pbSystemPolicies.SysPolicies = analyzer.GetSystemPolicies(in.GetSysLog())
	return &pbSystemPolicies, nil
}

// =================== //
// == Observability == //
// =================== //

type observabilityServer struct {
	opb.ObservabilityServer
}

func (s *observabilityServer) SysObservabilityData(ctx context.Context, in *opb.Data) (*opb.Response, error) {

	var wpfs types.WorkloadProcessFileSet

	wpfs.ContainerName = in.ContainerName
	wpfs.ClusterName = in.ClusterName
	wpfs.FromSource = in.FromSource
	wpfs.Namespace = in.Namespace
	wpfs.Labels = in.Labels

	if in.Request == "observe" {
		wpfs.FromSource = "" // Forcing wpfs.FromSource to "" as it is not a search string
		resp, err := sysworker.GetSystemObsData(wpfs)
		return &resp, err
	}
	if in.Request == "dbclear" {
		err := sysworker.ClearSysDb(wpfs, in.Duration)
		return &opb.Response{}, err
	}

	return nil, errors.New("not a valid request, use observe/dbclear")
}

// ================= //
// == gRPC server == //
// ================= //

func GetNewServer() *grpc.Server {
	s := grpc.NewServer()
	grpc_health_v1.RegisterHealthServer(s, health.NewServer())

	reflection.Register(s)

	// create server instances
	workerServer := &workerServer{}
	consumerServer := &consumerServer{}
	analyzerServer := &analyzerServer{}
	observabilityServer := &observabilityServer{}

	// register gRPC servers
	wpb.RegisterWorkerServer(s, workerServer)
	fpb.RegisterConsumerServer(s, consumerServer)
	apb.RegisterAnalyzerServer(s, analyzerServer)
	opb.RegisterObservabilityServer(s, observabilityServer)

	if cfg.GetCurrentCfg().ConfigClusterMgmt.ClusterInfoFrom != "k8sclient" {
		// start consumer automatically
		feedconsumer.StartConsumer()
	}

	// start net worker automatically
	networker.StartNetworkWorker()

	// start sys worker automatically
	sysworker.StartSystemWorker()

	return s
}
