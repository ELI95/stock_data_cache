package cache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const defaultBasePath = "/cache/"
const Sina = "sina"

// HTTPPool implements PeerPicker for a pool of HTTP peers.
type HTTPPool struct {
	// this peer's base URL, e.g. "https://example.net:8000"
	self     string
	basePath string
}

// NewHTTPPool initializes an HTTP pool of peers.
func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

// Log info with server name
func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP handle all http requests
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE, UPDATE")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")

	// /<basepath>/<groupname>?key=...
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		http.Error(w, "unexpected path: "+r.URL.Path, http.StatusNotFound)
	}
	p.Log("%s %s", r.Method, r.URL.Path)

	if strings.Contains(r.URL.Path[len(p.basePath):], "/") {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	groupName := r.URL.Path[len(p.basePath):]

	keys, ok := r.URL.Query()["key"]
	if !ok || len(keys[0]) < 1 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	key := keys[0]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	if values, ok := r.URL.Query()["value"]; ok {
		// update cache
		value := values[0]
		group.populateCache(key, ByteView{b: cloneBytes([]byte(value))})
	}

	// get cache
	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, err = w.Write(view.ByteSlice())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
