package w32

import (
	"testing"
	"fmt"
)

func TestLastInputInfo(t *testing.T) {
	info := GetLastInputInfo()
	fmt.Println(info)

}
