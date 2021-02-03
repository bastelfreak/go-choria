package puppetsec

import (
	"fmt"
	"os"
	"runtime"

	"github.com/choria-io/go-choria/tlssetup"

	"github.com/choria-io/go-choria/config"
	"github.com/sirupsen/logrus"
)

// Option is a function that can configure the Puppet Security Provider
type Option func(*PuppetSecurity) error

// WithChoriaConfig optionally configures the Puppet Security Provider from settings found in a typical Choria configuration
func WithChoriaConfig(c *config.Config) Option {
	return func(p *PuppetSecurity) error {
		cfg := Config{
			AllowList:                    c.Choria.CertnameWhitelist,
			DisableTLSVerify:             c.DisableTLSVerify,
			PrivilegedUsers:              c.Choria.PrivilegedUsers,
			SSLDir:                       c.Choria.SSLDir,
			PuppetCAHost:                 c.Choria.PuppetCAHost,
			PuppetCAPort:                 c.Choria.PuppetCAPort,
			Identity:                     c.Identity,
			AlwaysOverwriteCache:         c.Choria.SecurityAlwaysOverwriteCache,
			RemoteSignerURL:              c.Choria.RemoteSignerURL,
			RemoteSignerTokenFile:        c.Choria.RemoteSignerTokenFile,
			RemoteSignerTokenEnvironment: c.Choria.RemoteSignerTokenEnvironment,
			TLSConfig:                    tlssetup.TLSConfig(c),
			noLegacyTLS:                  c.InitiatedBySystem && !c.InitiatedByServer,
		}

		if c.Choria.NetworkClientAdvertiseName != "" {
			cfg.AltNames = append(cfg.AltNames, c.Choria.NetworkClientAdvertiseName)
		}

		if c.HasOption("plugin.choria.puppetca_host") || c.HasOption("plugin.choria.puppetca_port") {
			cfg.DisableSRV = true
		}

		if c.OverrideCertname == "" {
			if cn, ok := os.LookupEnv("MCOLLECTIVE_CERTNAME"); ok {
				c.OverrideCertname = cn
			}
		}

		if c.OverrideCertname != "" {
			cfg.Identity = c.OverrideCertname
		} else if !c.InitiatedByServer && !c.InitiatedBySystem {
			userEnvVar := "USER"

			if runtime.GOOS == "windows" {
				userEnvVar = "USERNAME"
			}

			u, ok := os.LookupEnv(userEnvVar)
			if ok {
				cfg.Identity = fmt.Sprintf("%s.mcollective", u)
			}
		}

		if cfg.SSLDir == "" {
			d, err := userSSlDir()
			if err != nil {
				return err
			}

			cfg.SSLDir = d
		}

		p.conf = &cfg

		return nil
	}
}

// WithConfig optionally configures the Puppet Security Provider using its native configuration format
func WithConfig(c *Config) Option {
	return func(p *PuppetSecurity) error {
		p.conf = c

		return nil
	}
}

// WithLog configures a logger for the Puppet Security Provider
func WithLog(l *logrus.Entry) Option {
	return func(p *PuppetSecurity) error {
		p.log = l.WithFields(logrus.Fields{"ssl": "puppet"})

		return nil
	}
}

// WithResolver configures a SRV resolver for the Puppet Security Provider
func WithResolver(r Resolver) Option {
	return func(p *PuppetSecurity) error {
		p.res = r

		return nil
	}
}
