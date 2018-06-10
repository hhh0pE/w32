package w32

import (
	"testing"
	"fmt"
)

func TestQueryFullProcessImageName(t *testing.T) {
	procHandle, err := OpenProcess(PROCESS_QUERY_INFORMATION, false, 1492)
	fmt.Println(procHandle, err)

	QueryFullProcessImageName(procHandle, 0)

	//syscall.OpenProcess()
}
