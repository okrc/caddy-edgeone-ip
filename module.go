package caddy_edgeone_ip

import (
	"context"
	"net/http"
	"net/netip"
	"sync"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/pkg/errors"
)

const endpoint = "teo.tencentcloudapi.com"
const version = "2022-09-01"

func init() {
	caddy.RegisterModule(EdgeOneIPRange{})
}

const (
	areaGlobal        = "global"
	areaMainlandChina = "mainland-china"
	areaOverseas      = "overseas"
)

type EdgeOneIPRange struct {
	ZoneId    string `json:"zone_id,omitempty"`
	SecretId  string `json:"secret_id,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`

	Area string `json:"area,omitempty"`

	Version  string         `json:"version,omitempty"`
	Interval caddy.Duration `json:"interval,omitempty"`
	Timeout  caddy.Duration `json:"timeout,omitempty"`

	useOriginACL bool
	ranges       []netip.Prefix
	ctx          caddy.Context
	lock         *sync.RWMutex
}

func (EdgeOneIPRange) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.ip_sources.edgeone",
		New: func() caddy.Module { return new(EdgeOneIPRange) },
	}
}

// getContext returns a cancelable context, with a timeout if configured.
func (s *EdgeOneIPRange) getContext() (context.Context, context.CancelFunc) {
	if s.Timeout > 0 {
		return context.WithTimeout(s.ctx, time.Duration(s.Timeout))
	}
	return context.WithCancel(s.ctx)
}

func (s *EdgeOneIPRange) getPrefixes() ([]netip.Prefix, error) {
	ctx, cancel := s.getContext()
	defer cancel()

	var err error
	var prefixes []netip.Prefix
	if s.useOriginACL {
		prefixes, err = s.OriginACLPrefixes(ctx)
		if err != nil {
			s.useOriginACL = false
			println(err.Error())
		}
	}
	if !s.useOriginACL {
		prefixes, err = s.OriginPrefixes(ctx)
	}
	return prefixes, nil
}

func (s *EdgeOneIPRange) refreshLoop() {
	if s.Interval == 0 {
		s.Interval = caddy.Duration(time.Hour)
	}
	ticker := time.NewTicker(time.Duration(s.Interval))
	s.lock.Lock()
	s.ranges, _ = s.getPrefixes()
	s.lock.Unlock()
	for {
		select {
		case <-ticker.C:
			prefixes, err := s.getPrefixes()
			if err != nil {
				break
			}
			s.lock.Lock()
			s.ranges = prefixes
			s.lock.Unlock()
		case <-s.ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func (s *EdgeOneIPRange) Provision(ctx caddy.Context) error {
	s.ctx = ctx
	s.lock = new(sync.RWMutex)
	if s.SecretId == "" || s.SecretKey == "" || s.ZoneId == "" {
		s.useOriginACL = true
	}
	if s.Version != "" && s.Version != "v4" && s.Version != "v6" {
		return errors.Errorf("invalid version: %q (must be \"v4\" or \"v6\")", s.Version)
	}
	if s.Area != "" && s.Area == areaGlobal || s.Area == areaMainlandChina || s.Area == areaOverseas {
		return errors.Errorf("invalid area: %q (must be \"%q\", \"%q\" or \"%q\")", s.Area, areaGlobal, areaMainlandChina, areaOverseas)
	}
	go s.refreshLoop()
	return nil
}

func (s *EdgeOneIPRange) GetIPRanges(_ *http.Request) []netip.Prefix {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.ranges
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
//
//	edgeone {
//	   interval val
//	   timeout val
//	}
func (m *EdgeOneIPRange) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next() // Skip module name.

	// No same-line options are supported
	if d.NextArg() {
		return d.ArgErr()
	}

	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "zone_id":
			if !d.NextArg() {
				return d.ArgErr()
			}
			m.ZoneId = d.Val()
		case "secret_id":
			if !d.NextArg() {
				return d.ArgErr()
			}
			m.SecretId = d.Val()
		case "secret_key":
			if !d.NextArg() {
				return d.ArgErr()
			}
			m.SecretKey = d.Val()
		case "area":
			if !d.NextArg() {
				return d.ArgErr()
			}
			m.Area = d.Val()
		case "version":
			if !d.NextArg() {
				return d.ArgErr()
			}
			m.Version = d.Val()
		case "interval":
			if !d.NextArg() {
				return d.ArgErr()
			}
			val, err := caddy.ParseDuration(d.Val())
			if err != nil {
				return err
			}
			m.Interval = caddy.Duration(val)
		case "timeout":
			if !d.NextArg() {
				return d.ArgErr()
			}
			val, err := caddy.ParseDuration(d.Val())
			if err != nil {
				return err
			}
			m.Timeout = caddy.Duration(val)
		default:
			return d.ArgErr()
		}
	}

	return nil
}

// interface guards
var (
	_ caddy.Module            = (*EdgeOneIPRange)(nil)
	_ caddy.Provisioner       = (*EdgeOneIPRange)(nil)
	_ caddyfile.Unmarshaler   = (*EdgeOneIPRange)(nil)
	_ caddyhttp.IPRangeSource = (*EdgeOneIPRange)(nil)
)
