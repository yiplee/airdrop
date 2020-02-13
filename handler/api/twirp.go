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
	"github.com/twitchtv/twirp"
)

// custom code in error meta
const CustomCode = "custom_code"

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
	if opt.TransformFn == nil {
		opt.TransformFn = defaultTransformerFn
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

type twirpErr struct {
	Code string            `json:"code,omitempty"`
	Msg  string            `json:"msg,omitempty"`
	Meta map[string]string `json:"meta,omitempty"`
}

func (err twirpErr) customCode() int {
	if err.Meta == nil {
		return 0
	}

	v, ok := err.Meta[CustomCode]
	if !ok {
		return 0
	}

	code, _ := strconv.Atoi(v)
	return code
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

func (w *wrapResponse) writeData(body []byte) (int, error) {
	b, err := json.Marshal(struct {
		Data json.RawMessage `json:"data,omitempty"`
	}{Data: body})
	if err != nil {
		return 0, err
	}

	return w.finalWrite(b)
}

func (w *wrapResponse) writeErr(body []byte) (int, error) {
	var twerr twirpErr
	_ = json.Unmarshal(body, &twerr)

	var r struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Hint string `json:"hint,omitempty"`
	}

	if code := twerr.customCode(); code > 0 {
		r.Code = code
		r.Msg = twerr.Msg
	} else {
		r.Code = twirp.ServerHTTPStatusFromErrorCode(twirp.ErrorCode(twerr.Code))
		r.Msg = http.StatusText(r.Code)
	}

	if ResponseErrorMessageAsHint {
		r.Hint = twerr.Msg
	}

	b, err := json.Marshal(r)
	if err != nil {
		return 0, err
	}

	return w.finalWrite(b)
}
