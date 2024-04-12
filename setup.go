package serverip

import (
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

func init() { plugin.Register("serverip", setup) }

func setup(c *caddy.Controller) error {
	c.Next() // 'serverip'
	if c.NextArg() {
		return plugin.Error("serverip", c.ArgErr())
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return Serverip{}
	})

	return nil
}
