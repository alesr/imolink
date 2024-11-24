package apierror

import (
	"encore.dev/beta/errs"
	"encore.dev/rlog"
)

func E(msg string, e error, code errs.ErrCode) error {
	if e == nil {
		return nil
	}
	rlog.Error(msg, "error", e)
	return &errs.Error{
		Code:    code,
		Message: msg,
	}
}
