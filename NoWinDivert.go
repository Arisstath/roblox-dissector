// +build !divert

package main

import "errors"

const WinDivertEnabled = false

func (session *CaptureSession) CaptureFromWinDivert() error {
	// nop
	return errors.New("windivert disabled at build time")
}
