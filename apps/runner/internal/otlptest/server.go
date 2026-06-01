// Copyright Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

// Package otlptest provides an in-process OTLP/HTTP collector useful for
// asserting that the runner emits the metrics, traces, and logs we expect
// over the wire. It is intended for tests only — the package name and
// internal/ placement keep it out of the runner binary.
//
// The server accepts the three OTLP/HTTP signal endpoints with binary
// protobuf (Content-Type: application/x-protobuf) and stores decoded payloads
// in memory for later inspection. Gzipped bodies are transparently decoded.
//
// Usage:
//
//	srv := otlptest.New(t)
//	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", srv.URL)
//	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")
//	t.Setenv("OTEL_EXPORTER_OTLP_INSECURE", "true")
//	... configure otel.Init, emit metrics, shutdown to flush ...
//	srv.AssertCounterValue(t, "docker.registry.operation.count", map[string]string{...}, 1)
package otlptest

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"

	collectorlogs "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	collectormetrics "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	collectortrace "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonv1 "go.opentelemetry.io/proto/otlp/common/v1"
	logsv1 "go.opentelemetry.io/proto/otlp/logs/v1"
	metricsv1 "go.opentelemetry.io/proto/otlp/metrics/v1"
	tracev1 "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/protobuf/proto"
)

// Server is an in-process OTLP/HTTP receiver.
type Server struct {
	*httptest.Server

	mu      sync.Mutex
	metrics []*metricsv1.ResourceMetrics
	traces  []*tracev1.ResourceSpans
	logs    []*logsv1.ResourceLogs
}

// New starts an OTLP/HTTP server on a random localhost port and registers
// cleanup to stop it at test end.
func New(t testing.TB) *Server {
	t.Helper()
	s := &Server{}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/metrics", s.handleMetrics)
	mux.HandleFunc("/v1/traces", s.handleTraces)
	mux.HandleFunc("/v1/logs", s.handleLogs)

	s.Server = httptest.NewServer(mux)
	t.Cleanup(s.Close)
	return s
}

// Reset clears all received signals. Useful between subtests.
func (s *Server) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metrics = nil
	s.traces = nil
	s.logs = nil
}

// ResourceMetrics returns a snapshot of all metric resource batches received
// so far. The returned slice is safe to inspect without further locking.
func (s *Server) ResourceMetrics() []*metricsv1.ResourceMetrics {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*metricsv1.ResourceMetrics, len(s.metrics))
	copy(out, s.metrics)
	return out
}

// ResourceSpans returns a snapshot of all span resource batches.
func (s *Server) ResourceSpans() []*tracev1.ResourceSpans {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*tracev1.ResourceSpans, len(s.traces))
	copy(out, s.traces)
	return out
}

// ResourceLogs returns a snapshot of all log resource batches.
func (s *Server) ResourceLogs() []*logsv1.ResourceLogs {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*logsv1.ResourceLogs, len(s.logs))
	copy(out, s.logs)
	return out
}

// ── signal handlers ─────────────────────────────────────────────────────────

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	body, err := readBody(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var req collectormetrics.ExportMetricsServiceRequest
	if err := proto.Unmarshal(body, &req); err != nil {
		http.Error(w, fmt.Sprintf("unmarshal metrics: %v", err), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.metrics = append(s.metrics, req.GetResourceMetrics()...)
	s.mu.Unlock()

	writeProto(w, &collectormetrics.ExportMetricsServiceResponse{})
}

func (s *Server) handleTraces(w http.ResponseWriter, r *http.Request) {
	body, err := readBody(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var req collectortrace.ExportTraceServiceRequest
	if err := proto.Unmarshal(body, &req); err != nil {
		http.Error(w, fmt.Sprintf("unmarshal traces: %v", err), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.traces = append(s.traces, req.GetResourceSpans()...)
	s.mu.Unlock()

	writeProto(w, &collectortrace.ExportTraceServiceResponse{})
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	body, err := readBody(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var req collectorlogs.ExportLogsServiceRequest
	if err := proto.Unmarshal(body, &req); err != nil {
		http.Error(w, fmt.Sprintf("unmarshal logs: %v", err), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.logs = append(s.logs, req.GetResourceLogs()...)
	s.mu.Unlock()

	writeProto(w, &collectorlogs.ExportLogsServiceResponse{})
}

// ── lookup helpers ──────────────────────────────────────────────────────────

// AttrsMatch reports whether every attribute in want is present in got with
// the matching string value. Attributes in got that are not listed in want
// are ignored.
func AttrsMatch(got []*commonv1.KeyValue, want map[string]string) bool {
	for k, v := range want {
		var found bool
		for _, kv := range got {
			if kv.GetKey() != k {
				continue
			}
			if kv.GetValue().GetStringValue() != v {
				return false
			}
			found = true
			break
		}
		if !found {
			return false
		}
	}
	return true
}

// HasAttr reports whether attrs contains a key with any value.
func HasAttr(attrs []*commonv1.KeyValue, key string) bool {
	for _, kv := range attrs {
		if kv.GetKey() == key {
			return true
		}
	}
	return false
}

// SumCounter returns the cumulative value of the matching Int64 Sum data
// point summed across all received batches. Returns (0, false) if no matching
// point was seen. Data points are matched by name AND by requiring every
// attribute in want to be present with the matching value (extras allowed).
func (s *Server) SumCounter(name string, want map[string]string) (int64, bool) {
	var total int64
	var found bool
	for _, rm := range s.ResourceMetrics() {
		for _, sm := range rm.GetScopeMetrics() {
			for _, m := range sm.GetMetrics() {
				if m.GetName() != name {
					continue
				}
				sum := m.GetSum()
				if sum == nil {
					continue
				}
				for _, dp := range sum.GetDataPoints() {
					if !AttrsMatch(dp.GetAttributes(), want) {
						continue
					}
					total += dp.GetAsInt()
					found = true
				}
			}
		}
	}
	return total, found
}

// HistogramCount returns the count of histogram observations matching name +
// attribute filter, summed across all received batches.
func (s *Server) HistogramCount(name string, want map[string]string) (uint64, bool) {
	var total uint64
	var found bool
	for _, rm := range s.ResourceMetrics() {
		for _, sm := range rm.GetScopeMetrics() {
			for _, m := range sm.GetMetrics() {
				if m.GetName() != name {
					continue
				}
				hist := m.GetHistogram()
				if hist == nil {
					continue
				}
				for _, dp := range hist.GetDataPoints() {
					if !AttrsMatch(dp.GetAttributes(), want) {
						continue
					}
					total += dp.GetCount()
					found = true
				}
			}
		}
	}
	return total, found
}

// ResourceAttrs returns the merged resource attributes across all received
// metric batches. If the runner emits multiple batches with the same resource,
// later batches override earlier ones for the same key.
func (s *Server) ResourceAttrs() map[string]string {
	out := map[string]string{}
	for _, rm := range s.ResourceMetrics() {
		for _, kv := range rm.GetResource().GetAttributes() {
			out[kv.GetKey()] = kv.GetValue().GetStringValue()
		}
	}
	return out
}

// MetricNames returns the de-duplicated set of metric instrument names across
// all received batches.
func (s *Server) MetricNames() []string {
	seen := map[string]struct{}{}
	for _, rm := range s.ResourceMetrics() {
		for _, sm := range rm.GetScopeMetrics() {
			for _, m := range sm.GetMetrics() {
				seen[m.GetName()] = struct{}{}
			}
		}
	}
	names := make([]string, 0, len(seen))
	for n := range seen {
		names = append(names, n)
	}
	return names
}

// AnyDataPointMissing returns true if no data point with the given metric name
// and attribute filter has the attribute key in absentKey set.
// In other words: assert that none of the matching points carries a particular
// attribute (e.g. confirm sandbox.id is omitted when caller passed nil).
func (s *Server) AnyDataPointMissing(name string, want map[string]string, absentKey string) bool {
	for _, rm := range s.ResourceMetrics() {
		for _, sm := range rm.GetScopeMetrics() {
			for _, m := range sm.GetMetrics() {
				if m.GetName() != name {
					continue
				}
				for _, dp := range allDataPointAttrs(m) {
					if !AttrsMatch(dp, want) {
						continue
					}
					if HasAttr(dp, absentKey) {
						return false
					}
				}
			}
		}
	}
	return true
}

// Dump writes a human-readable view of every received metric to t.Log so it
// shows up under `go test -v`. Useful while developing instrumentation to see
// exactly what the runner ships over the wire.
//
// Layout:
//
//	== OTLP collector dump ==
//	resource:
//	  service.name=daytona-runner
//	  service.namespace=runner
//	  ...
//	metrics:
//	  docker.registry.operation.count [scope=github.com/daytonaio/runner/docker]
//	    SUM[int] {operation=pull, registry.host=cr.example.com, sandbox.id=sb-1, status=success} = 1
//	    SUM[int] {operation=pull, registry.host=docker.io, status=success} = 1
//	  docker.registry.operation.duration [scope=...]
//	    HIST count=1 sum=0.052s {operation=pull, ...}
//
// Resource attributes are de-duplicated across batches.
func (s *Server) Dump(t testing.TB) {
	t.Helper()
	var b strings.Builder
	b.WriteString("== OTLP collector dump ==\n")

	res := s.ResourceAttrs()
	if len(res) > 0 {
		b.WriteString("resource:\n")
		for _, k := range sortedKeys(res) {
			fmt.Fprintf(&b, "  %s=%s\n", k, res[k])
		}
	}

	b.WriteString("metrics:\n")

	type instrumentKey struct {
		scope string
		name  string
	}
	// Group data points by (scope, name) and preserve receive order roughly via
	// first-seen scope.
	grouped := map[instrumentKey][]string{}
	order := []instrumentKey{}

	for _, rm := range s.ResourceMetrics() {
		for _, sm := range rm.GetScopeMetrics() {
			scope := sm.GetScope().GetName()
			for _, m := range sm.GetMetrics() {
				key := instrumentKey{scope: scope, name: m.GetName()}
				if _, ok := grouped[key]; !ok {
					order = append(order, key)
				}
				grouped[key] = append(grouped[key], formatMetric(m)...)
			}
		}
	}

	for _, key := range order {
		fmt.Fprintf(&b, "  %s [scope=%s]\n", key.name, key.scope)
		for _, line := range grouped[key] {
			fmt.Fprintf(&b, "    %s\n", line)
		}
	}

	if len(order) == 0 {
		b.WriteString("  (none received)\n")
	}

	t.Log(b.String())
}

func formatMetric(m *metricsv1.Metric) []string {
	var lines []string
	if sum := m.GetSum(); sum != nil {
		for _, dp := range sum.GetDataPoints() {
			value := "?"
			switch v := dp.GetValue().(type) {
			case *metricsv1.NumberDataPoint_AsInt:
				value = fmt.Sprintf("%d", v.AsInt)
			case *metricsv1.NumberDataPoint_AsDouble:
				value = fmt.Sprintf("%g", v.AsDouble)
			}
			lines = append(lines, fmt.Sprintf("SUM %s %s = %s", attrString(dp.GetAttributes()), monotonicTag(sum.GetIsMonotonic()), value))
		}
	}
	if hist := m.GetHistogram(); hist != nil {
		for _, dp := range hist.GetDataPoints() {
			lines = append(lines, fmt.Sprintf("HIST count=%d sum=%g %s", dp.GetCount(), dp.GetSum(), attrString(dp.GetAttributes())))
		}
	}
	if g := m.GetGauge(); g != nil {
		for _, dp := range g.GetDataPoints() {
			value := "?"
			switch v := dp.GetValue().(type) {
			case *metricsv1.NumberDataPoint_AsInt:
				value = fmt.Sprintf("%d", v.AsInt)
			case *metricsv1.NumberDataPoint_AsDouble:
				value = fmt.Sprintf("%g", v.AsDouble)
			}
			lines = append(lines, fmt.Sprintf("GAUGE %s = %s", attrString(dp.GetAttributes()), value))
		}
	}
	return lines
}

func attrString(attrs []*commonv1.KeyValue) string {
	if len(attrs) == 0 {
		return "{}"
	}
	pairs := make([]string, 0, len(attrs))
	for _, kv := range attrs {
		pairs = append(pairs, fmt.Sprintf("%s=%s", kv.GetKey(), kv.GetValue().GetStringValue()))
	}
	sort.Strings(pairs)
	return "{" + strings.Join(pairs, ", ") + "}"
}

func monotonicTag(monotonic bool) string {
	if monotonic {
		return "(counter)"
	}
	return "(updown)"
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func allDataPointAttrs(m *metricsv1.Metric) [][]*commonv1.KeyValue {
	var out [][]*commonv1.KeyValue
	if sum := m.GetSum(); sum != nil {
		for _, dp := range sum.GetDataPoints() {
			out = append(out, dp.GetAttributes())
		}
	}
	if hist := m.GetHistogram(); hist != nil {
		for _, dp := range hist.GetDataPoints() {
			out = append(out, dp.GetAttributes())
		}
	}
	if g := m.GetGauge(); g != nil {
		for _, dp := range g.GetDataPoints() {
			out = append(out, dp.GetAttributes())
		}
	}
	return out
}

// ── HTTP helpers ────────────────────────────────────────────────────────────

func readBody(r *http.Request) ([]byte, error) {
	defer r.Body.Close()

	var reader io.Reader = r.Body
	if r.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			return nil, fmt.Errorf("gzip reader: %w", err)
		}
		defer gz.Close()
		reader = gz
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	return body, nil
}

func writeProto(w http.ResponseWriter, msg proto.Message) {
	out, err := proto.Marshal(msg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/x-protobuf")
	_, _ = w.Write(out)
}

// errors that downstream packages might want to identify.
var (
	// ErrNotFound indicates a metric data point matching the filter was not
	// received. It is currently unused but exported so the test package can
	// produce stable wrapped errors if needed in future iterations.
	ErrNotFound = errors.New("otlptest: metric data point not found")
)
