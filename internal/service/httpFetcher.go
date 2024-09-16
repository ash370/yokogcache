package service

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type httpFetcher struct {
	baseURL string //要访问的远程节点的地址，such "http://localhost:9999/_yokogcache/"
}

var _ Fetcher = (*httpFetcher)(nil)

func (h *httpFetcher) Fetch(group string, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	) //最终得到地址格式应该是：/<basepath>/<groupname>/<key>，这是服务端能够解析的地址格式

	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() //确保在函数返回前关闭响应体 res.Body，避免资源泄漏

	if res.StatusCode != http.StatusOK { //检查响应的 HTTP 状态码
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := io.ReadAll(res.Body) //读取响应体数据
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}
