// Copyright 2010-2012 The W32 Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package w32

import (
	"syscall"
	"unsafe"
	"log"
)

var (
	powerProf = syscall.NewLazyDLL("powrprof.dll")

	callNtPowerInformationProc         = powerProf.NewProc("CallNtPowerInformation")

)

func CallNtPowerInformation(level uint, inputBuffer uintptr, inputBufferLength int, outputBuffer uintptr, outputBufferLength int) bool {
	r, _, lastErr := callNtPowerInformationProc.Call(uintptr(level), uintptr(inputBuffer), uintptr(inputBufferLength), uintptr(outputBuffer), uintptr(outputBufferLength))
	if lastErr != nil {
		log.Println("CallNtPowerInformation error: "+lastErr.Error())
	}
	return r != 0
}

type ThreadExecutionState int
const (
	ES_AWAYMODE_REQUIRED ThreadExecutionState = 0x00000040
	ES_CONTINUOUS = 0x80000000
	ES_DISPLAY_REQUIRED = 0x00000002
	ES_SYSTEM_REQUIRED = 0x00000001
	ES_USER_PRESENT = 0x00000004
)

func GetSystemExecutionState(execState *uint) bool {
	var powerResult int
	return CallNtPowerInformation(16, 0, 0, uintptr(unsafe.Pointer(execState)), int(unsafe.Sizeof(powerResult)))
}