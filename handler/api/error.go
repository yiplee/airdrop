package api

import (
	"strconv"

	"github.com/twitchtv/twirp"
)

// custom code in error meta
const CustomCode = "custom_code"

type twirpErr struct {
	Code twirp.ErrorCode   `json:"code,omitempty"`
	Msg  string            `json:"msg,omitempty"`
	Meta map[string]string `json:"meta,omitempty"`
}

func (err twirpErr) meta(key string) string {
	if err.Meta == nil {
		return ""
	}

	return err.Meta[key]
}

func (err twirpErr) displayCode() int {
	m := err.meta(CustomCode)
	if m == "" {
		switch err.Code {
		case twirp.InvalidArgument:
			m = InvalidArgument
		}
	}

	if code, _ := strconv.Atoi(m); code > 0 {
		return code
	}

	return twirp.ServerHTTPStatusFromErrorCode(err.Code)
}

func (err twirpErr) displayMsg() string {
	// InternalErrorWith 产生的 error 会在 meta 带一个 cause
	// 替换掉 msg 避免将内部错误信息暴露给前端
	if cause := err.meta("cause"); cause != "" {
		return string(err.Code)
	}

	return err.Msg
}
