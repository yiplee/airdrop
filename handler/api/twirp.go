package api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
)

// Response internal error msg as hint
var ResponseErrorMessageAsHint bool

// Twirp modify the request and redirect to rpc handler
func Twirp(handler http.Handler, opt TwirpOpts) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		opt.repackRequest(r)

		if routeCtx := chi.RouteContext(r.Context()); routeCtx != nil {
			routeCtx.Reset()
		}

		w = &wrapResponse{
			writer: w,
			header: http.Header{},
		}
		handler.ServeHTTP(w, r)
	}
}

type TwirpOpts struct {
	PathPrefix string
	Method     string
	// TransformFn transform url params and query items
	TransformFn func(key, value string) interface{}
}

func defaultTransformFn(_, value string) interface{} { return value }

func (opt TwirpOpts) repackRequest(r *http.Request) {
	if opt.TransformFn == nil {
		opt.TransformFn = defaultTransformFn
	}

	forms := make(map[string]interface{})
	if routeCtx := chi.RouteContext(r.Context()); routeCtx != nil {
		params := routeCtx.URLParams
		for idx, key := range params.Keys {
			forms[key] = opt.TransformFn(key, params.Values[idx])
		}
	}

	for key, items := range r.URL.Query() {
		value := strings.Join(items, ",")
		forms[key] = opt.TransformFn(key, value)
	}

	if len(forms) > 0 {
		_ = json.NewDecoder(r.Body).Decode(&forms)
		_ = r.Body.Close()
		b := &bytes.Buffer{}
		_ = json.NewEncoder(b).Encode(forms)
		r.Body = ioutil.NopCloser(b)
	}

	r.Method = http.MethodPost
	r.URL.RawQuery = ""
	r.URL.Path = path.Join(opt.PathPrefix, opt.Method)
	r.Header.Set("Content-Type", "application/json")
}

type wrapResponse struct {
	writer http.ResponseWriter
	header http.Header
	status int
}

func (w *wrapResponse) Header() http.Header {
	return w.header
}

func (w *wrapResponse) WriteHeader(statusCode int) {
	w.status = statusCode
}

func (w *wrapResponse) finalWrite(body []byte) (int, error) {
	// reset content length
	w.header.Set("Content-Length", strconv.Itoa(len(body)))

	for key := range w.header {
		w.writer.Header().Set(key, w.header.Get(key))
	}

	w.writer.WriteHeader(w.status)
	return w.writer.Write(body)
}

func (w *wrapResponse) Write(body []byte) (int, error) {
	switch w.status {
	case http.StatusOK:
		return w.writeData(body)
	default:
		return w.writeErr(body)
	}
}

type dataResponse struct {
	Data json.RawMessage `json:"data,omitempty"`
}

func (w *wrapResponse) writeData(body []byte) (int, error) {
	r := dataResponse{Data: body}
	b, err := json.Marshal(r)
	if err != nil {
		return 0, err
	}

	return w.finalWrite(b)
}

type errorResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Hint string `json:"hint,omitempty"`
}

func (w *wrapResponse) writeErr(body []byte) (int, error) {
	var twerr twirpErr
	_ = json.Unmarshal(body, &twerr)

	r := errorResponse{
		Code: twerr.displayCode(),
		Msg:  twerr.displayMsg(),
	}

	if ResponseErrorMessageAsHint && r.Msg != twerr.Msg {
		r.Hint = twerr.Msg
	}

	b, err := json.Marshal(r)
	if err != nil {
		return 0, err
	}

	return w.finalWrite(b)
}
