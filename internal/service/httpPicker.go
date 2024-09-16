package service

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"yokogcache/internal/service/consistenthash"
)

var _ PeerPicker = (*HTTPPool)(nil)

type HTTPPool struct {
	//peer's base URL, "http://example.net:8000"
	self         string //like localhost:9999
	basePath     string //默认"/_yokogcache/"作为前缀
	ring         *consistenthash.ConsistentHash
	mu           sync.Mutex
	httpFetchers map[string]*httpFetcher
}

// 给所有节点初始化一个HTTP池子
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// 日志打印时加上服务器名称
func (h *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", h.self, fmt.Sprintf(format, v...))
}

// 实现Handler接口，服务端监听到的请求会进入这里被处理
func (h *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//检查：如果前缀不是节点间通讯地址前缀，直接报错
	if !strings.HasPrefix(r.URL.Path, h.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}

	h.Log("%s %s", r.Method, r.URL.Path)

	//地址格式要求：/<basepath>/<groupname>/<key>
	parts := strings.SplitN(r.URL.Path[len(h.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupname := parts[0]
	key := parts[1]

	group := groups[groupname]
	if group == nil {
		http.Error(w, "no such group: "+groupname, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice()) //返回数据的深拷贝，不能返回原始数据
}

func (h *HTTPPool) Pick(key string) (Fetcher, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	peerAddr := h.ring.GetTruthNode(key)
	if peerAddr == "" || peerAddr == h.self {
		// upper layer get the value of the key locally after receiving false
		return nil, false
	}
	h.Log("Pick peer %s", peerAddr)
	return h.httpFetchers[peerAddr], true
}

func (h *HTTPPool) UpdatePeers(peers ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.ring = consistenthash.NewConsistentHash(defaultReplicas, nil)
	h.ring.AddTruthNodes(peers...)
	h.httpFetchers = map[string]*httpFetcher{}

	for _, peer := range peers {
		h.httpFetchers[peer] = &httpFetcher{
			baseURL: peer + h.basePath, // such "http://localhost:9999/_yokogcache/"
		}
	}
}
