package proxy

import "runtime"

// Outputter is what is used by the tracing proxy created via `NewTraceProxy`.
// Anything that implements a `log.Logger` style `Output` method will satisfy
// this interface.
type Outputter interface {
	Output(calldepth int, s string) error
}

// Filter is used by the tracing proxy for skipping database libraries (e.g. O/R mapper).
type Filter interface {
	DoOutput(packageName string) bool
}

// PackageFilter is an implementation of Filter.
type PackageFilter map[string]struct{}

// DoOutput returns false if the package is in the ignored list.
func (f PackageFilter) DoOutput(packageName string) bool {
	_, ok := f[packageName]
	return !ok
}

// Ignore add the package into the ignored list.
func (f PackageFilter) Ignore(packageName string) {
	f[packageName] = struct{}{}
}

// DefaultPackageFilter ignores some database util package.
var DefaultPackageFilter = PackageFilter{
	"database/sql":                       struct{}{},
	"github.com/shogo82148/txmanager":    struct{}{},
	"github.com/shogo82148/go-sql-proxy": struct{}{},
}

func findCaller(f Filter) int {
	// skip starts 4. 0: Callers, 1: findCaller, 2: hooks, 3: proxy-funcs, 4: database/sql, and equals or greater than 5: user-funcs
	skip := 5
	for {
		var rpc [8]uintptr
		n := runtime.Callers(skip, rpc[:])

		for i, pc := range rpc[:n] {
			// http://stackoverflow.com/questions/25262754/how-to-get-name-of-current-package-in-go
			name := runtime.FuncForPC(pc).Name()
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
				// -1 because the meaning of skip differs between Caller and Callers.
				return skip + i - 1
			}
		}
		if n < len(rpc) {
			break
		}
		skip += n
	}
	return 0
}
