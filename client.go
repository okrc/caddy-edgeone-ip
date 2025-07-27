package edgeone

import (
	"context"
	"io"
	"net/http"
	"strings"
)

func (s *EdgeOneIPRange) doAPIRequest(ctx context.Context, action string, data string) ([]byte, error) {
	endpointUrl := "https://teo.tencentcloudapi.com"
	req, err := http.NewRequestWithContext(ctx, "POST", endpointUrl, strings.NewReader(data))
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
