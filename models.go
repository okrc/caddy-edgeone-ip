package caddy_edgeone_ip

import "encoding/json"

type DescribeOriginACLResponse struct {
	Response struct {
		Error         *Error        `json:"Error,omitempty"`
		OriginACLInfo OriginACLInfo `json:"OriginACLInfo,omitempty"`
		RequestId     string        `json:"RequestId,omitempty"`
	}
}

type OriginACLInfo struct {
	L7Hosts          []string         `json:"L7Hosts,omitempty"`
	L4ProxyIds       []string         `json:"L4ProxyIds,omitempty"`
	CurrentOriginACL CurrentOriginACL `json:"CurrentOriginACL,omitempty"`
	NextOriginACL    NextOriginACL    `json:"NextOriginACL,omitempty"`
	Status           string           `json:"Status,omitempty"`
}

type CurrentOriginACL struct {
	EntireAddresses Addresses `json:"EntireAddresses,omitempty"`
	Version         string    `json:"Version,omitempty"`
	ActiveTime      string    `json:"ActiveTime,omitempty"`
	IsPlaned        string    `json:"IsPlaned,omitempty"`
}

type NextOriginACL struct {
	Version           string    `json:"Version,omitempty"`
	PlannedActiveTime string    `json:"PlannedActiveTime,omitempty"`
	EntireAddresses   Addresses `json:"EntireAddresses,omitempty"`
	AddedAddresses    Addresses `json:"AddedAddresses,omitempty"`
	RemovedAddresses  Addresses `json:"RemovedAddresses,omitempty"`
	NoChangeAddresses Addresses `json:"NoChangeAddresses,omitempty"`
}

type Addresses struct {
	IPv4 []string `json:"IPv4,omitempty"`
	IPv6 []string `json:"IPv6,omitempty"`
}

func (r *DescribeOriginACLResponse) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

func (r *DescribeOriginACLResponse) FromJsonString(s []byte) error {
	return json.Unmarshal(s, &r)
}

type DescribeOriginACLRequest struct {
	ZoneId string `json:"ZoneId,omitempty"`
}

func (r *DescribeOriginACLRequest) ToJsonString() string {
	b, _ := json.Marshal(r)
	return string(b)
}

type Error struct {
	Code    string
	Message string
}
