// Package logger provides a colorful, human-friendly slog logger for
// development and a structured JSON logger for production.
package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// New returns a *slog.Logger. In the "development" environment it writes
// pretty, colorized lines to stderr; otherwise it emits JSON.
func New(env string) *slog.Logger {
	if env == "development" {
		color := os.Getenv("NO_COLOR") == ""
		return slog.New(newPrettyHandler(os.Stderr, slog.LevelDebug, color))
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
}

// ── ANSI colors ─────────────────────────────────────────────────────────────

const (
	reset   = "\x1b[0m"
	bold    = "\x1b[1m"
	dim     = "\x1b[2m"
	red     = "\x1b[31m"
	green   = "\x1b[32m"
	yellow  = "\x1b[33m"
	blue    = "\x1b[34m"
	magenta = "\x1b[35m"
	cyan    = "\x1b[36m"
	gray    = "\x1b[90m"
)

// prettyHandler renders slog records as aligned, colorized console lines.
// HTTP access logs (msg "REQUEST"/"REQUEST_ERROR" emitted by echo's
// RequestLogger) get a compact, status-colored layout.
type prettyHandler struct {
	mu    *sync.Mutex
	w     io.Writer
	level slog.Leveler
	color bool
	attrs []slog.Attr
	group string
}

func newPrettyHandler(w io.Writer, level slog.Leveler, color bool) *prettyHandler {
	return &prettyHandler{mu: &sync.Mutex{}, w: w, level: level, color: color}
}

func (h *prettyHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *prettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clone := *h
	clone.attrs = append(append([]slog.Attr{}, h.attrs...), attrs...)
	return &clone
}

func (h *prettyHandler) WithGroup(name string) slog.Handler {
	clone := *h
	if h.group != "" {
		clone.group = h.group + "." + name
	} else {
		clone.group = name
	}
	return &clone
}

func (h *prettyHandler) paint(color, s string) string {
	if !h.color {
		return s
	}
	return color + s + reset
}

func (h *prettyHandler) Handle(_ context.Context, r slog.Record) error {
	// Collect attributes (record + handler-bound) into a map for lookup.
	attrs := make(map[string]slog.Value, r.NumAttrs()+len(h.attrs))
	order := make([]string, 0, r.NumAttrs())
	for _, a := range h.attrs {
		attrs[a.Key] = a.Value
	}
	r.Attrs(func(a slog.Attr) bool {
		if _, seen := attrs[a.Key]; !seen {
			order = append(order, a.Key)
		}
		attrs[a.Key] = a.Value
		return true
	})

	ts := h.paint(gray, r.Time.Format("15:04:05.000"))

	var line string
	if r.Message == "REQUEST" || r.Message == "REQUEST_ERROR" {
		line = h.formatRequest(ts, attrs)
	} else {
		line = h.formatGeneric(ts, r, attrs, order)
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := io.WriteString(h.w, line+"\n")
	return err
}

// formatRequest renders the compact HTTP access-log line.
func (h *prettyHandler) formatRequest(ts string, attrs map[string]slog.Value) string {
	method := fmt.Sprintf("%-6s", strVal(attrs, "method"))
	uri := strVal(attrs, "uri")
	status := int(durSafeInt(attrs, "status"))
	latency := durSafe(attrs, "latency")

	var b strings.Builder
	b.WriteString(ts)
	b.WriteByte(' ')
	b.WriteString(h.paint(methodColor(strings.TrimSpace(method)), bold+method+reset))
	b.WriteByte(' ')
	b.WriteString(h.paint(statusColor(status), bold+strconv.Itoa(status)+reset))
	b.WriteByte(' ')
	b.WriteString(uri)
	b.WriteString("  ")
	b.WriteString(h.paint(dim, fmtLatency(latency)))

	if errVal, ok := attrs["error"]; ok {
		b.WriteString("  ")
		b.WriteString(h.paint(red, "→ "+errVal.String()))
	}
	return b.String()
}

// formatGeneric renders any other log record.
func (h *prettyHandler) formatGeneric(ts string, r slog.Record, attrs map[string]slog.Value, order []string) string {
	var b strings.Builder
	b.WriteString(ts)
	b.WriteByte(' ')
	b.WriteString(h.paint(levelColor(r.Level), levelLabel(r.Level)))
	b.WriteByte(' ')
	b.WriteString(r.Message)

	// Hidden internal attrs from the request logger we don't want in generic lines.
	skip := map[string]bool{}

	keys := order
	if len(h.attrs) > 0 {
		seen := map[string]bool{}
		for _, k := range order {
			seen[k] = true
		}
		bound := make([]string, 0, len(h.attrs))
		for _, a := range h.attrs {
			if !seen[a.Key] {
				bound = append(bound, a.Key)
			}
		}
		sort.Strings(bound)
		keys = append(bound, order...)
	}

	for _, k := range keys {
		if skip[k] {
			continue
		}
		v := attrs[k]
		b.WriteByte(' ')
		b.WriteString(h.paint(cyan, k))
		b.WriteByte('=')
		b.WriteString(v.String())
	}
	return b.String()
}

// ── helpers ─────────────────────────────────────────────────────────────────

func strVal(attrs map[string]slog.Value, key string) string {
	if v, ok := attrs[key]; ok {
		return v.String()
	}
	return ""
}

// durSafeInt extracts an integer attribute without panicking on kind mismatch.
func durSafeInt(attrs map[string]slog.Value, key string) int64 {
	v, ok := attrs[key]
	if !ok {
		return 0
	}
	if v.Kind() == slog.KindInt64 {
		return v.Int64()
	}
	return 0
}

// durSafe extracts a duration attribute without panicking on kind mismatch.
func durSafe(attrs map[string]slog.Value, key string) time.Duration {
	v, ok := attrs[key]
	if !ok {
		return 0
	}
	if v.Kind() == slog.KindDuration {
		return v.Duration()
	}
	return 0
}

func levelLabel(l slog.Level) string {
	switch {
	case l >= slog.LevelError:
		return "ERR"
	case l >= slog.LevelWarn:
		return "WRN"
	case l >= slog.LevelInfo:
		return "INF"
	default:
		return "DBG"
	}
}

func levelColor(l slog.Level) string {
	switch {
	case l >= slog.LevelError:
		return red
	case l >= slog.LevelWarn:
		return yellow
	case l >= slog.LevelInfo:
		return green
	default:
		return gray
	}
}

func statusColor(code int) string {
	switch {
	case code >= 500:
		return red
	case code >= 400:
		return yellow
	case code >= 300:
		return cyan
	case code >= 200:
		return green
	default:
		return gray
	}
}

func methodColor(method string) string {
	switch method {
	case "GET":
		return blue
	case "POST":
		return green
	case "PUT", "PATCH":
		return yellow
	case "DELETE":
		return red
	default:
		return magenta
	}
}

func fmtLatency(d time.Duration) string {
	switch {
	case d < time.Microsecond:
		return strconv.FormatInt(d.Nanoseconds(), 10) + "ns"
	case d < time.Millisecond:
		return strconv.FormatFloat(float64(d.Nanoseconds())/1e3, 'f', 1, 64) + "µs"
	case d < time.Second:
		return strconv.FormatFloat(float64(d.Nanoseconds())/1e6, 'f', 1, 64) + "ms"
	default:
		return strconv.FormatFloat(d.Seconds(), 'f', 2, 64) + "s"
	}
}
