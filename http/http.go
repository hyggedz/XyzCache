package XyzCache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const defaultBasePath = "/_xyzcache/"

// HTTPPool 管理不同节点之间的通信
type HTTPPool struct {
	self     string
	basePath string
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
	segs := strings.Split(r.URL.Path[len(p.basePath):], "/")
	if len(segs) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := segs[0]
	key := segs[1]

	g := GetGroup(groupName)
	if g == nil {
		http.Error(w, "group not found", http.StatusNotFound)
	}

	bv, err := g.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(bv.ByteSlice())
}
