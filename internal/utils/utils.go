package utils

import (
	"fmt"
	"os"
)

// KernelPanic prints the error and triggers a kernel panic.
//
// This function WILL cause a system crash!
// DO NOT call it in any code that may be covered by automatic or manual tests.
func KernelPanic(err error) {
	fmt.Println(err)
	_ = os.WriteFile("/proc/sys/kernel/sysrq", []byte("1"), 0o644)
	_ = os.WriteFile("/proc/sysrq-trigger", []byte("c"), 0o644)
	panic(err)
}
