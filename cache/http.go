package cache

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

const defaultBasePath = "/cache/"
const Sina = "sina"

type UpdateCacheRequest struct {
	Key string `json:"key"`
	Value string `json:"value"`
}

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
	fmt.Printf("receive request, method: %s, path: %s\n", r.Method, r.URL.Path)
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		http.Error(w, "unexpected path: "+r.URL.Path, http.StatusNotFound)
	}

	if strings.Contains(r.URL.Path[len(p.basePath):], "/") {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	groupName := r.URL.Path[len(p.basePath):]
	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	if _, ok := r.URL.Query()["missed"]; ok {
		// get missed
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		select {
		case key := <- group.mainCache.missedChan:
			fmt.Printf("missed channel length: %d\n", len(group.mainCache.missedChan))
			_, err := w.Write([]byte(key))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		default:
		}
		w.WriteHeader(200)
		return
	}

	if r.Method == "POST" {
		// update cache
		var params UpdateCacheRequest
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&params); err != nil {
			fmt.Printf("update cache failed, error: %s\n", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Printf("update cache succeed, key: %s, value: %s\n", params.Key, params.Value)
		group.populateCache(params.Key, ByteView{b: cloneBytes([]byte(params.Value))})
		w.WriteHeader(200)
		return
	}

	// get cache
	keys, ok := r.URL.Query()["key"]
	if !ok || len(keys[0]) < 1 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	key := keys[0]

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
	w.WriteHeader(200)
}
