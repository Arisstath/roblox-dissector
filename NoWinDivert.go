// +build !divert

package main

import (
	"context"
	"errors"
)

func CaptureFromDivert(_ context.Context, session *CaptureSession) error {
	session.ReportDone()
	return errors.New("windivert disabled at build time")
}
