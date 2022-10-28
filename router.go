package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

type BaseRouter struct {
	key         string
	path        string
	len         int                    // path length, use to split
	handlerFunc http.Handler           `json:"-"`
	children    map[string]*BaseRouter `json:"-"`
}

func NewBaseRouter() *BaseRouter {
	return &BaseRouter{
		children: make(map[string]*BaseRouter),
	}
}

// GetKey returns the string key
func (b *BaseRouter) GetKey() string {
	return b.key
}

func (b *BaseRouter) String() string {
	return fmt.Sprintf("key:%s,path:%s,len:%d", b.key, b.path, b.len)
}

// WithKey returns a copy with the given string key
func (b *BaseRouter) WithKey(key string) *BaseRouter {
	b.key = key
	return b
}

// GetPath returns the string path
func (b *BaseRouter) GetPath() string {
	return b.path
}

// WithPath returns a copy with the given string path
func (b *BaseRouter) WithPath(path string) *BaseRouter {
	b.path = path
	return b
}

// GetLen returns the int len
func (b *BaseRouter) GetLen() int {
	return b.len
}

// WithLen returns a copy with the given int len
func (b *BaseRouter) WithLen(len int) *BaseRouter {
	b.len = len
	return b
}

// GetHandlerFunc returns the http.Handler handlerFunc
func (b *BaseRouter) GetHandlerFunc() http.Handler {
	return b.handlerFunc
}

// IsNoFunc returns the http.Handler handlerFunc is nil
func (b *BaseRouter) IsNoFunc() bool {
	return b.handlerFunc == nil
}

// WithHandlerFunc returns a copy with the given http.Handler handlerFunc
func (b *BaseRouter) WithHandlerFunc(handlerFunc http.Handler) *BaseRouter {
	b.handlerFunc = handlerFunc
	return b
}

// GetChildrenKeys returns the string keys of children of type *BaseRouter
func (b *BaseRouter) GetChildrenKeys() []string {
	keys := make([]string, len(b.children))
	i := 0
	for k := range b.children {
		keys[i] = k
		i++
	}
	return keys
}

// ChildrenAt returns the children of type *BaseRouter at the requested key
func (b *BaseRouter) ChildrenAt(key string) *BaseRouter {
	return b.children[key]
}

// WithChildren returns a copy with the *BaseRouter with the given key string
func (b *BaseRouter) WithChildren(key string, value *BaseRouter) *BaseRouter {
	b.children[key] = value
	return b
}

// //////////////////////////////////////////////////////////////

const splitString = "/"

func (b *BaseRouter) AddRouter(path string, hand http.Handler) {
	b.addRouter(path, path, hand)
}

func (b *BaseRouter) addRouter(path, allPath string, hand http.Handler) {
	if strings.HasPrefix(path, splitString) {
		path = path[1:]
	}
	if path == "" {
		b.WithPath(allPath).
			WithLen(len(allPath)).
			WithHandlerFunc(hand)
		return
	}
	index := strings.Index(path, splitString)
	if index < 0 {
		children := b.ChildrenAt(path)
		if children == nil {
			children = NewBaseRouter().
				WithHandlerFunc(hand).
				WithKey(path).
				WithPath(allPath).
				WithLen(len(allPath))
			b.WithChildren(path, children)
			return
		} else {
			children.WithHandlerFunc(hand).
				WithKey(path).
				WithPath(allPath).
				WithLen(len(allPath))
			return
		}
	} else {
		key := path[:index]
		children := b.ChildrenAt(key)
		if children == nil {
			children = NewBaseRouter().WithKey(key)
			b.WithChildren(key, children)
			children.addRouter(path[index+1:], allPath, hand)
		} else {
			children.WithKey(key).addRouter(path[index+1:], allPath, hand)
		}
	}

}

func (b *BaseRouter) CheckRouter(path string) *BaseRouter {
	if strings.HasPrefix(path, splitString) {
		path = path[1:]
	}
	if path == "" {
		return b
	}
	index := strings.Index(path, splitString)
	if index < 0 {
		children := b.ChildrenAt(path)
		if children == nil {
			// /1/2/3/4 --> /1/2/3 匹配的数据为最长的那个
			return b
		} else {
			return children
		}
	} else {
		key := path[:index]
		children := b.ChildrenAt(key)
		if children == nil {
			// /1/2/3/4/ --> /1/2/3 匹配的数据为最长的那个
			return b
		} else {
			return children.CheckRouter(path[index+1:])
		}
	}
}

func (b *BaseRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	hand := b.CheckRouter(path)
	if hand == nil || hand.IsNoFunc() {
		w.WriteHeader(404)
		return
	}
	hand.GetHandlerFunc().ServeHTTP(w, r)
}

// ////////////////////////

var err404 = []byte(`{"err":"404"}`)
var errHttp = func(err error) []byte {
	if err == nil {
		return []byte(`{"err":""}`)
	}
	return []byte(fmt.Sprintf(`{"err":"%s"}`, err.Error()))
}

func HttpHandler(router string, to string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, router)
		var buf strings.Builder
		buf.WriteString(to)
		buf.WriteString(path)
		if r.URL.ForceQuery || r.URL.RawQuery != "" {
			buf.WriteByte('?')
			buf.WriteString(r.URL.RawQuery)
		}
		u := buf.String()
		_http(u, w, r)
	}
}

func FileHandler(router string, _path string) http.HandlerFunc {
	server := http.FileServer(http.Dir(_path))
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, router)
		r.URL.Path = path
		log.Info().Str("to", r.URL.String()).Str("router", router).Msg("file")
		server.ServeHTTP(w, r)
	}
}

func _http(u string, w http.ResponseWriter, r *http.Request) {

	log.Info().Str("input", r.URL.String()).Str("to", u).Msg("http")

	defer r.Body.Close()
	req, err := http.NewRequest(r.Method, u, r.Body)
	if err != nil {
		w.WriteHeader(500)
		w.Write(errHttp(err))
		return
	}
	req.Header = r.Header
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		w.WriteHeader(500)
		w.Write(errHttp(err))
		return
	}
	w.WriteHeader(resp.StatusCode)
	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}
	defer resp.Body.Close()
	io.Copy(w, resp.Body)
}
