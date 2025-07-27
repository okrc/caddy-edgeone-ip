package edgeone

import (
	"context"
	"errors"
	"net/http"
	"net/netip"
	"sync"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

const endpoint = "teo.tencentcloudapi.com"
const version = "2022-09-01"

func init() {
	caddy.RegisterModule(EdgeOneIPRange{})
}

type EdgeOneIPRange struct {
	ZoneId    string         `json:"zone_id"`
	SecretId  string         `json:"secret_id"`
	SecretKey string         `json:"secret_key"`
	Interval  caddy.Duration `json:"interval,omitempty"`
	Timeout   caddy.Duration `json:"timeout,omitempty"`

	ranges []netip.Prefix
	ctx    caddy.Context
	lock   *sync.RWMutex
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

	requestData := DescribeOriginACLRequest{ZoneId: s.ZoneId}
	resp, err := s.doAPIRequest(ctx, "DescribeOriginACL", requestData.ToJsonString())
	if err != nil {
		return nil, err
	}

	var response DescribeOriginACLResponse
	if err := response.FromJsonString(resp); err != nil {
		return nil, err
	}
	if response.Response.Error != nil {
		return nil, errors.New(response.Response.Error.Message)
	}
	var prefixes []netip.Prefix
	for _, ip := range append(response.Response.OriginACLInfo.CurrentOriginACL.EntireAddresses.IPv4,
		response.Response.OriginACLInfo.CurrentOriginACL.EntireAddresses.IPv6...) {
		prefix, err := caddyhttp.CIDRExpressionToPrefix(ip)
		if err != nil {
			return nil, err
		}
		prefixes = append(prefixes, prefix)
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
			if d.NextArg() {
				m.ZoneId = d.Val()
			}
		case "secret_id":
			if d.NextArg() {
				m.SecretId = d.Val()
			}
		case "secret_key":
			if d.NextArg() {
				m.SecretKey = d.Val()
			}
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

	if m.SecretId == "" || m.SecretKey == "" || m.ZoneId == "" {
		return d.Err("missing required field: zone_id, secret_id, or secret_key")
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
