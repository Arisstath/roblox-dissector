// +build !divert

package main

const WinDivertEnabled = false

func (session *CaptureSession) CaptureFromDivert(_, _ string) {
	// nop
}
