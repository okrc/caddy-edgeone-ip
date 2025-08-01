package caddy_edgeone_ip

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net/http"
	"net/netip"
	"net/url"
	"strings"

	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func (s *EdgeOneIPRange) doAPIRequest(ctx context.Context, action string, data string) ([]byte, error) {
	endpointUrl := "https://teo.tencentcloudapi.com"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpointUrl, strings.NewReader(data))
	if err != nil {
		return nil, err
	}
	tencentCloudSigner(s.SecretId, s.SecretKey, req, action, data, edgeOneService)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func (s *EdgeOneIPRange) OriginACLPrefixes(ctx context.Context) ([]netip.Prefix, error) {
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
	var fullPrefixes []netip.Prefix
	if s.Version != "v6" {
		prefixes, err := s.getIPs(response.Response.OriginACLInfo.CurrentOriginACL.EntireAddresses.IPv4)
		if err != nil {
			return nil, err
		}
		fullPrefixes = append(fullPrefixes, prefixes...)
	}
	if s.Version != "v4" {
		prefixes, err := s.getIPs(response.Response.OriginACLInfo.CurrentOriginACL.EntireAddresses.IPv6)
		if err != nil {
			return nil, err
		}
		fullPrefixes = append(fullPrefixes, prefixes...)
	}
	return fullPrefixes, nil
}

func (s *EdgeOneIPRange) OriginPrefixes(ctx context.Context) ([]netip.Prefix, error) {
	endpointUrl, err := url.Parse("https://api.edgeone.ai/ips")
	q := endpointUrl.Query()
	if s.Version == "v4" || s.Version == "v6" {
		q.Set("version", s.Version)
	}
	if s.Area == "global" || s.Area == "mainland-china" || s.Area == "overseas" {
		q.Set("area", s.Area)
	}
	endpointUrl.RawQuery = q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpointUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	var prefixes []netip.Prefix
	for scanner.Scan() {
		prefix, err := caddyhttp.CIDRExpressionToPrefix(scanner.Text())
		if err != nil {
			return nil, err
		}
		prefixes = append(prefixes, prefix)
	}
	return prefixes, nil
}

func (s *EdgeOneIPRange) getIPs(ips []string) ([]netip.Prefix, error) {
	var prefixes []netip.Prefix
	for _, ip := range ips {
		prefix, err := caddyhttp.CIDRExpressionToPrefix(ip)
		if err != nil {
			return nil, err
		}
		prefixes = append(prefixes, prefix)
	}
	return prefixes, nil
}
