//go:build go1.12
// +build go1.12

package proxy

import (
	"runtime"
	"strings"
)

func findCaller(f Filter) int {
	// skip starts 5. 0: Callers, 1: findCaller, 2: hooks, 3: proxy-funcs, 4: database/sql, and equals or greater than 5: user-funcs
	skip := 5
	for {
		var rpc [8]uintptr
		var i int
		n := runtime.Callers(skip, rpc[:])
		frames := runtime.CallersFrames(rpc[:])
		for i = 0; ; i++ {
			frame, more := frames.Next()
			if !more {
				break
			}
			name := frame.Function
			if name == "" || strings.HasPrefix(name, "runtime.") {
				continue
			}
			// http://stackoverflow.com/questions/25262754/how-to-get-name-of-current-package-in-go
			dotIdx := 0
			for j := len(name) - 1; j >= 0; j-- {
				if name[j] == '.' {
					dotIdx = j
				} else if name[j] == '/' {
					break
				}
			}
			packageName := name[:dotIdx]
			if f.DoOutput(packageName) {
				return skip + i
			}
		}
		if n < len(rpc) {
			break
		}
		skip += i
	}
	// fallback to the caller
	// 1: Outputter.Output, 2: the caller
	return 2
}
