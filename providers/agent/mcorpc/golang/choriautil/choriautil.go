package choriautil

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"github.com/choria-io/go-choria/aagent"
	"github.com/choria-io/go-choria/inter"
	"github.com/choria-io/go-choria/internal/util"
	"github.com/choria-io/go-choria/protocol"
	"github.com/choria-io/go-choria/providers/agent/mcorpc"
	"github.com/choria-io/go-choria/server"
	"github.com/choria-io/go-choria/server/agents"
	"github.com/nats-io/nats.go"
)

type info struct {
	Security          string   `json:"security"`
	Connector         string   `json:"connector"`
	ClientVersion     string   `json:"client_version"`
	ClientFlavour     string   `json:"client_flavour"`
	ClientOptions     *copts   `json:"client_options"`
	ClientStats       *cstats  `json:"client_stats"`
	ConnectedServer   string   `json:"connected_server"`
	FacterDomain      string   `json:"facter_domain"`
	FacterCommand     string   `json:"facter_command"`
	SrvDomain         string   `json:"srv_domain"`
	UsingSrv          bool     `json:"using_srv"`
	MiddlewareServers []string `json:"middleware_servers"`
	Path              string   `json:"path"`
	ChoriaVersion     string   `json:"choria_version"`
	ProtocolSecure    bool     `json:"secure_protocol"`
	ConnectorTLS      bool     `json:"connector_tls"`
}

type copts struct {
	Servers        []string `json:"servers"`
	NoRandomize    bool     `json:"dont_randomize_servers"`
	Name           string   `json:"name"`
	Pedantic       bool     `json:"pedantic"`
	Secure         bool     `json:"secure"`
	AllowReconnect bool     `json:"allow_reconnect"`
	MaxReconnect   int      `json:"max_reconnect_attempts"`
	ReconnectWait  float64  `json:"reconnect_time_wait"`
	Timeout        float64  `json:"connect_timeout"`
	PingInterval   float64  `json:"ping_interval"`
	MaxPingsOut    int      `json:"max_outstanding_pings"`
}

type cstats struct {
	InMsgs     uint64 `json:"in_msgs"`
	OutMsgs    uint64 `json:"out_msgs"`
	InBytes    uint64 `json:"in_bytes"`
	OutBytes   uint64 `json:"out_bytes"`
	Reconnects uint64 `json:"reconnects"`
}

type machineStates struct {
	MachineIDs   []string                       `json:"machine_ids"`
	MachineNames []string                       `json:"machine_names"`
	States       map[string]aagent.MachineState `json:"states"`
}

type machineTransitionRequest struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	ID         string `json:"instance"`
	Path       string `json:"path"`
	Transition string `json:"transition"`
}

type machineTransitionReply struct {
	Success bool `json:"success"`
}

type machineStateRequest struct {
	Name string `json:"name"`
	ID   string `json:"instance"`
	Path string `json:"path"`
}

type machineStateResponse struct {
	aagent.MachineState
}

// New creates a new choria_util agent
func New(mgr server.AgentManager) (*mcorpc.Agent, error) {
	bi := util.BuildInfo()

	metadata := &agents.Metadata{
		Name:        "choria_util",
		Description: "Choria Utilities",
		Author:      "R.I.Pienaar <rip@devco.net>",
		Version:     bi.Version(),
		License:     bi.License(),
		Timeout:     10,
		URL:         "https://choria.io",
	}

	agent := mcorpc.New("choria_util", metadata, mgr.Choria(), mgr.Logger())

	agent.MustRegisterAction("info", infoAction)
	agent.MustRegisterAction("machine_states", machineStatesAction)
	agent.MustRegisterAction("machine_state", machineStateAction)
	agent.MustRegisterAction("machine_transition", machineTransitionAction)

	return agent, nil
}

func machineStateAction(ctx context.Context, req *mcorpc.Request, reply *mcorpc.Reply, agent *mcorpc.Agent, conn inter.ConnectorInfo) {
	i := machineStateRequest{}
	if !mcorpc.ParseRequestData(&i, req, reply) {
		return
	}

	output := &machineStateResponse{
		aagent.MachineState{Name: "unknown machine", State: "unknown machine", Version: "unknown machine"},
	}
	reply.Data = output

	if i.Name == "" && i.Path == "" && i.ID == "" {
		reply.Statuscode = mcorpc.Aborted
		reply.Statusmsg = "No search criteria given"
		return
	}

	states, err := agent.ServerInfoSource.MachinesStatus()
	if err != nil {
		reply.Statuscode = mcorpc.Aborted
		reply.Statusmsg = fmt.Sprintf("Failed to retrieve states: %s", err)
		return
	}

	if len(states) == 0 {
		reply.Statuscode = mcorpc.Aborted
		reply.Statusmsg = "No running machines"
		return
	}

	var found []aagent.MachineState
	for _, state := range states {
		nameMatch := i.Name == ""
		pathMatch := i.Path == ""
		idMatch := i.ID == ""

		if i.Name != "" {
			nameMatch = state.Name == i.Name
		}

		if i.Path != "" {
			pathMatch = state.Path == i.Path
		}

		if i.ID != "" {
			idMatch = state.ID == i.ID
		}

		if nameMatch && pathMatch && idMatch {
			found = append(found, state)
		}
	}

	if len(found) > 1 {
		reply.Statuscode = mcorpc.Aborted
		reply.Statusmsg = "Found multiple machines matching criteria"
		return
	}

	if len(found) == 0 {
		reply.Statuscode = mcorpc.Aborted
		reply.Statusmsg = "Could not find a matching machine"
		return
	}

	output.MachineState = found[0]
}

func machineTransitionAction(ctx context.Context, req *mcorpc.Request, reply *mcorpc.Reply, agent *mcorpc.Agent, conn inter.ConnectorInfo) {
	i := machineTransitionRequest{}
	if !mcorpc.ParseRequestData(&i, req, reply) {
		return
	}

	err := agent.ServerInfoSource.MachineTransition(i.Name, i.Version, i.Path, i.ID, i.Transition)
	if err != nil {
		reply.Statuscode = mcorpc.Aborted
		reply.Statusmsg = fmt.Sprintf("Could not perform %s transition: %s", i.Transition, err)
	}

	reply.Data = machineTransitionReply{Success: err == nil}
}

func machineStatesAction(ctx context.Context, req *mcorpc.Request, reply *mcorpc.Reply, agent *mcorpc.Agent, conn inter.ConnectorInfo) {
	states, err := agent.ServerInfoSource.MachinesStatus()
	if err != nil {
		reply.Statuscode = mcorpc.Aborted
		reply.Statusmsg = fmt.Sprintf("Failed to retrieve states: %s", err)
		return
	}

	r := machineStates{
		MachineIDs:   []string{},
		MachineNames: []string{},
		States:       make(map[string]aagent.MachineState),
	}

	for _, m := range states {
		r.MachineIDs = append(r.MachineIDs, m.ID)
		r.MachineNames = append(r.MachineNames, fmt.Sprintf("%s %s", m.Name, m.Version))

		r.States[m.ID] = m
	}

	reply.Data = r
}

func infoAction(ctx context.Context, req *mcorpc.Request, reply *mcorpc.Reply, agent *mcorpc.Agent, conn inter.ConnectorInfo) {
	c := agent.Config

	domain, err := agent.Choria.FacterDomain()
	if err != nil {
		domain = ""
	}

	servers, err := agent.Choria.MiddlewareServers()
	if err != nil {
		reply.Statuscode = mcorpc.Aborted
		reply.Statusmsg = fmt.Sprintf("Could not determine middleware servers: %s", err)
	}

	mservers := servers.HostPorts()
	options := conn.ConnectionOptions()
	stats := conn.ConnectionStats()
	bi := util.BuildInfo()

	reply.Data = &info{
		Security:          "choria",
		Connector:         "choria",
		ClientVersion:     nats.Version,
		ClientFlavour:     fmt.Sprintf("nats.go %s", runtime.Version()),
		ConnectedServer:   conn.ConnectedServer(),
		FacterCommand:     agent.Choria.FacterCmd(),
		FacterDomain:      domain,
		SrvDomain:         c.Choria.SRVDomain,
		MiddlewareServers: mservers,
		Path:              os.Getenv("PATH"),
		ChoriaVersion:     fmt.Sprintf("choria %s", bi.Version()),
		UsingSrv:          c.Choria.UseSRVRecords,
		ProtocolSecure:    protocol.IsSecure(),
		ConnectorTLS:      bi.HasTLS(),
		ClientStats: &cstats{
			InMsgs:     stats.InMsgs,
			InBytes:    stats.InBytes,
			OutMsgs:    stats.OutMsgs,
			OutBytes:   stats.OutBytes,
			Reconnects: stats.Reconnects,
		},
		ClientOptions: &copts{
			Servers:        options.Servers,
			NoRandomize:    options.NoRandomize,
			Name:           options.Name,
			Pedantic:       options.Pedantic,
			Secure:         options.Secure,
			AllowReconnect: options.AllowReconnect,
			MaxReconnect:   options.MaxReconnect,
			ReconnectWait:  options.ReconnectWait.Seconds(),
			Timeout:        options.Timeout.Seconds(),
			PingInterval:   options.PingInterval.Seconds(),
			MaxPingsOut:    options.MaxPingsOut,
		},
	}
}
