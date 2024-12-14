package xyzcache

import (
	"XyzCache/xyzcache/consistentHash"
	pb "XyzCache/xyzcache/xyzcachepb"
	"fmt"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_xyzcache/"
	defaultReplicas = 50
)

type httpGetter struct {
	baseURL string
}

func (h *httpGetter) Get(in *pb.Request, out *pb.Response) error {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(in.Group),
		url.QueryEscape(in.Key),
	)
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}

	if err = proto.Unmarshal(bytes, out); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}

	return nil
}

var _ PeerGetter = (*httpGetter)(nil)

// HTTPPool 管理不同节点之间的通信
type HTTPPool struct {
	self     string
	basePath string

	mu          sync.Mutex
	peers       *consistentHash.Map
	httpGetters map[string]*httpGetter
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP 处理所有req
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//if req.URL 不带 basePath bad req
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool: URL不合法")
	}

	p.Log("%s,%s", r.Method, r.URL.Path)

	//URL: basePath / groupName / key
	segs := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(segs) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := segs[0]
	key := segs[1]

	g := GetGroup(groupName)
	if g == nil {
		http.Error(w, "group not found", http.StatusNotFound)
		return
	}

	bv, err := g.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	body, err := proto.Marshal(&pb.Response{Value: bv.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body)
}

// Set 更新 peers
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.peers = consistentHash.NewMap(defaultReplicas, nil)
	p.peers.Add(peers...)

	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil)
