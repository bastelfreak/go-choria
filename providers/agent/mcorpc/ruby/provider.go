package ruby

import (
	"context"
	"fmt"

	"github.com/choria-io/go-choria/config"
	"github.com/choria-io/go-choria/inter"
	"github.com/choria-io/go-choria/providers/agent/mcorpc/ddl/agent"
	"github.com/choria-io/go-choria/server"
	"github.com/sirupsen/logrus"
)

// agents we do not ever wish to load from ruby
var denylist = []string{"rpcutil", "choria_util", "discovery"}

// Provider is a Agent Provider capable of executing old mcollective ruby agents
type Provider struct {
	cfg    *config.Config
	log    *logrus.Entry
	agents []*agent.DDL
}

// Initialize configures the agent provider
func (p *Provider) Initialize(cfg *config.Config, log *logrus.Entry) {
	p.cfg = cfg
	p.log = log.WithFields(logrus.Fields{"provider": "ruby"})

	p.loadAgents(p.cfg.Choria.RubyLibdir)
}

// RegisterAgents registers known ruby agents using a shimm agent
func (p *Provider) RegisterAgents(ctx context.Context, mgr server.AgentManager, connector inter.AgentConnector, log *logrus.Entry) error {
	for _, ddl := range p.Agents() {
		agent, err := NewRubyAgent(ddl, mgr)
		if err != nil {
			p.log.Errorf("Could not register Ruby agent %s: %s", agent.Name(), err)
			continue
		}

		err = mgr.RegisterAgent(ctx, agent.Name(), agent, connector)
		if err != nil {
			p.log.Errorf("Could not register Ruby agent %s: %s", agent.Name(), err)
			continue
		}
	}

	return nil
}

// Agents provides a list of loaded agent DDLs
func (p *Provider) Agents() []*agent.DDL {
	return p.agents
}

// Version reports the version for this provider
func (p *Provider) Version() string {
	return fmt.Sprintf("%s version %s", p.PluginName(), p.PluginVersion())
}
