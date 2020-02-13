package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path"
	"strings"

	"github.com/go-chi/chi"
)

func defaultTransformerFn(key, value string) interface{} {
	return value
}

type TwirpOpts struct {
	PathPrefix string
	Method     string
	// TransformFn transform url params and query items
	TransformFn func(key, value string) interface{}
}

func (opt TwirpOpts) repackRequest(r *http.Request) {
	data := make(map[string]interface{})
	_ = json.NewDecoder(r.Body).Decode(&data)
	_ = r.Body.Close()

	if opt.TransformFn == nil {
		opt.TransformFn = defaultTransformerFn
	}

	ctx := r.Context()
	if routeCtx := chi.RouteContext(ctx); routeCtx != nil {
		params := routeCtx.URLParams
		for idx, key := range params.Keys {
			data[key] = opt.TransformFn(key, params.Values[idx])
		}
	}

	for key, items := range r.URL.Query() {
		value := strings.Join(items, ",")
		data[key] = opt.TransformFn(key, value)
	}

	b := &bytes.Buffer{}
	_ = json.NewEncoder(b).Encode(data)

	r.Body = ioutil.NopCloser(b)
	r.Method = http.MethodPost
	r.URL.RawQuery = ""
	r.URL.Path = path.Join(opt.PathPrefix, opt.Method)
	r.Header.Set("Content-Type", "application/json")
}

// Twirp modify the request and redirect to rpc handler
func Twirp(handler http.Handler, opt TwirpOpts) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		opt.repackRequest(r)
		handler.ServeHTTP(w, r)
	}
}
