package apierror

import (
	"encore.dev/beta/errs"
	"encore.dev/rlog"
)

func E(msg string, e error) error {
	if e == nil {
		return nil
	}
	rlog.Error(msg, "error", e)
	return &errs.Error{
		Code:    errs.Internal,
		Message: msg,
	}
}
