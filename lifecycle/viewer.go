package lifecycle

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/choria-io/go-choria/inter"
	"github.com/choria-io/go-choria/srvcache"
	"github.com/sirupsen/logrus"
)

// SubscribeConnector is a connection to the middleware
type SubscribeConnector interface {
	QueueSubscribe(ctx context.Context, name string, subject string, group string, output chan inter.ConnectorMessage) error
	ConnectedServer() string
}

type Framework interface {
	NewConnector(ctx context.Context, servers func() (srvcache.Servers, error), name string, logger *logrus.Entry) (conn inter.Connector, err error)
	Certname() string
	Logger(name string) *logrus.Entry
	NewRequestID() (string, error)
	MiddlewareServers() (servers srvcache.Servers, err error)
}

// ViewOptions configure the view command
type ViewOptions struct {
	TypeFilter      string
	ComponentFilter string
	Debug           bool
	Output          io.Writer
	Choria          Framework
	Connector       SubscribeConnector
}

// View connects and stream events to Output
func View(ctx context.Context, opt *ViewOptions) error {
	var err error

	log := opt.Choria.Logger("event_viewer")
	opt.Connector, err = opt.Choria.NewConnector(ctx, opt.Choria.MiddlewareServers, opt.Choria.Certname(), log)
	if err != nil {
		return fmt.Errorf("cannot connect: %s", err)
	}

	if opt.Output == nil {
		opt.Output = os.Stdout
	}

	fmt.Fprintf(opt.Output, "Waiting for events from topic choria.lifecycle.event.> on %s\n", opt.Connector.ConnectedServer())

	return WriteEvents(ctx, opt)
}

// WriteEvents views the event stream to the output
func WriteEvents(ctx context.Context, opt *ViewOptions) error {
	events := make(chan inter.ConnectorMessage, 100)

	rid, err := opt.Choria.NewRequestID()
	if err != nil {
		return err
	}

	err = opt.Connector.QueueSubscribe(ctx, rid, "choria.lifecycle.event.>", "", events)
	if err != nil {
		return fmt.Errorf("could not subscribe to event source: %s", err)
	}

	for {
		select {
		case e := <-events:
			event, err := NewFromJSON(e.Data())
			if err != nil {
				continue
			}

			if opt.ComponentFilter != "" {
				if event.Component() != opt.ComponentFilter {
					continue
				}
			}

			if opt.TypeFilter != "" {
				if event.TypeString() != opt.TypeFilter {
					continue
				}
			}

			if opt.Debug {
				fmt.Fprintf(opt.Output, "%s\n", string(e.Data()))
				continue
			}

			fmt.Fprintf(opt.Output, "%s %s\n", time.Now().Format("15:04:05"), event.String())

		case <-ctx.Done():
			return nil
		}
	}
}
