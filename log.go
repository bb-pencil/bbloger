package bbloger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"log"
	"runtime"
	"sort"
	"strings"
)
// The global verbosity level.
var globalVerbosity int = 0

func SetVerbosity(v int) int {
	old := globalVerbosity
	globalVerbosity = v
	return old
}

type Logger interface {
	// Output is the same as log.Output and log.Logger.Output.
	Output(calldepth int, logline string) error
}

type bbloger struct {
	loger  Logger
	level  int
	depth  int
	values []interface{}
	prefix string
}


func New(l Logger) logr.Logger {
	return NewWithOptions(l, Options{})
}

func NewWithOptions(l Logger, opts Options) logr.Logger {
	if opts.Depth < 0 {
		opts.Depth = 0
	}

	return bbloger{
		loger: l,
		level: 0,
		prefix: "",
		values: nil,
		depth: 0,
	}
}

type Options struct {
	// DepthOffset biases the assumed number of call frames to the "true"
	// caller.  This is useful when the calling code calls a function which then
	// calls glogr (e.g. a logging shim to another API).  Values less than zero
	// will be treated as zero.
	Depth int
}

func (l bbloger) V(level int) logr.Logger {
	new := l.clone()
	new.level += level
	return new
}

func (l bbloger) clone() bbloger {
	new := l
	l.values = copySlice(l.values)
	return new
}

func copySlice(in []interface{}) []interface{} {
	new := make([]interface{}, len(in))
	copy(new, in)
	return new
}

func (l bbloger) output(calldepth int, s string) {
	depth := calldepth + 2 // offset for this adapter
	if l.loger != nil {
		_ = l.loger.Output(depth, s)
	} else {
		_ = log.Output(depth, s)
	}
}

func (l bbloger) Info(msg string, kvList ...interface{})  {
	if l.Enabled() {
		lvlStr := flatten("level", l.level)
		msgStr := flatten("msg", msg)
		fixedStr := flatten(l.values...)
		userStr := flatten(kvList...)
		l.output(framesToCaller() + l.depth, fmt.Sprintln(l.prefix, lvlStr, msgStr, fixedStr, userStr))
	}
}

func (l bbloger) Error(err error, msg string, kvList ...interface{})  {
	msgStr := flatten("msg", msg)
	var loggableErr interface{}
	if err != nil {
		loggableErr = err.Error()
	}
	errStr := flatten("error", loggableErr)
	fixedStr := flatten(l.values...)
	userStr := flatten(kvList...)
	l.output(framesToCaller() + l.depth, fmt.Sprintln(l.prefix, errStr, msgStr, fixedStr, userStr))
}

// WithName returns a new logr.Logger with the specified name appended.
// bbloger uses '/' characters to separte name elements. Callers should not pass '/'
// in the provided name string, but this library does not actually enforce that.
func (l bbloger) WithName(name string) logr.Logger {
	new := l.clone()
	if len(l.prefix) > 0 {
		new.prefix = l.prefix + "/"
	}
	new.prefix += name
	return new
}

func (l bbloger) WithValues(kvList ...interface{}) logr.Logger {
	new := l.clone()
	new.values = append(new.values, kvList...)
	return new
}

func (l bbloger) Enabled() bool {
	return globalVerbosity >= l.level
}


// Magic string for intermediate frames that we should ignore.
const autogeneratedFrameName = "<autogenerated>"

// Discover how many frames ne need to climb to find the caller.
// This approach was suggested by Ian Lance Taylor of the Go team, so it *should* be safe enough (famous last words).
func framesToCaller() int {
	//
	for i := 1; i < 3; i++ {
		_, file, _, _ := runtime.Caller(i + 1)
		if file != autogeneratedFrameName {
			return i
		}
	}
	return 1	// something went wrong, this is safe
}

func flatten(kvList ...interface{}) string {
	keys := make([]string, 0, len(kvList))
	vals := make(map[string]interface{}, len(kvList))
	for i := 0; i < len(kvList); i += 2 {
		k, ok := kvList[i].(string)
		if !ok {
			panic(fmt.Sprintf("key is not a string: %s", pretty(kvList[i])))
		}
		var v interface{}
		if i+1 < len(kvList) {
			v = kvList[i+1]
		}
		keys = append(keys, k)
		vals[k] = v
	}
	sort.Strings(keys)
	buf := bytes.Buffer{}
	for i, k := range keys {
		v := vals[k]
		if i > 0 {
			buf.WriteRune(' ')
		}
		buf.WriteString(pretty(k))
		buf.WriteString("=")
		buf.WriteString(pretty(v))
	}
	return buf.String()
}

func pretty(value interface{}) string {
	if err, ok := value.(error); ok {
		if _, ok := value.(json.Marshaler); !ok {
			value = err.Error()
		}
	}
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	encoder.Encode(value)
	return strings.TrimSpace(string(buffer.Bytes()))
}

var _ logr.Logger = bbloger{}