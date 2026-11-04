package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"sort"
	"strings"
	"time"
)

// Weights relative to the price of one uncached input token.
const (
	weightCacheWrite = 1.25
	weightCacheRead  = 0.1
	weightOutput     = 5.0
)

// inputDollarsPerMillion is the price of one million uncached input tokens by model
// family; the other token kinds follow from the weights above.
var inputDollarsPerMillion = map[string]float64{"opus": 5, "sonnet": 3, "haiku": 1}

type tokens struct {
	Input, CacheWrite, CacheRead, Output float64
}

func (u tokens) weighted() float64 {
	return u.Input + weightCacheWrite*u.CacheWrite + weightCacheRead*u.CacheRead + weightOutput*u.Output
}

type turn struct {
	model, category string
	tokens
}

type record struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Message   struct {
		ID      string          `json:"id"`
		Model   string          `json:"model"`
		Content json.RawMessage `json:"content"`
		Usage   struct {
			Input      float64 `json:"input_tokens"`
			CacheWrite float64 `json:"cache_creation_input_tokens"`
			CacheRead  float64 `json:"cache_read_input_tokens"`
			Output     float64 `json:"output_tokens"`
		} `json:"usage"`
	} `json:"message"`
}

type transcript struct {
	turns       []turn
	first, last time.Time
}

// readTranscript counts one turn per API message. Claude Code writes a message once per
// content block, each copy carrying the same usage, so copies are merged by message id.
func readTranscript(r io.Reader) (*transcript, error) {
	t := &transcript{}
	byID := map[string]*turn{}
	var order []string
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 1<<20), 64<<20)
	for line := 1; sc.Scan(); line++ {
		if len(strings.TrimSpace(sc.Text())) == 0 {
			continue
		}
		var rec record
		if err := json.Unmarshal(sc.Bytes(), &rec); err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		if at, err := time.Parse(time.RFC3339Nano, rec.Timestamp); err == nil {
			if t.first.IsZero() || at.Before(t.first) {
				t.first = at
			}
			if at.After(t.last) {
				t.last = at
			}
		}
		if rec.Type != "assistant" || rec.Message.ID == "" {
			continue
		}
		cur, ok := byID[rec.Message.ID]
		if !ok {
			cur = &turn{model: rec.Message.Model, category: "text"}
			byID[rec.Message.ID] = cur
			order = append(order, rec.Message.ID)
		}
		u := rec.Message.Usage
		cur.tokens = tokens{math.Max(cur.Input, u.Input), math.Max(cur.CacheWrite, u.CacheWrite), math.Max(cur.CacheRead, u.CacheRead), math.Max(cur.Output, u.Output)}
		if cur.category == "text" {
			if tool := firstTool(rec.Message.Content); tool != "" {
				cur.category = tool
			}
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	for _, id := range order {
		t.turns = append(t.turns, *byID[id])
	}
	if len(t.turns) == 0 {
		return nil, errors.New("no assistant message with usage in the transcript")
	}
	return t, nil
}

// firstTool is the tool a message calls first, lowercased: the category of the whole
// turn, since the context it rereads is paid once however many tools it calls.
func firstTool(content json.RawMessage) string {
	var blocks []struct{ Type, Name string }
	if json.Unmarshal(content, &blocks) != nil {
		return ""
	}
	for _, b := range blocks {
		if b.Type == "tool_use" && b.Name != "" {
			return strings.ToLower(b.Name)
		}
	}
	return ""
}

func price(model string) (float64, error) {
	for family, dollars := range inputDollarsPerMillion {
		if strings.Contains(model, family) {
			return dollars, nil
		}
	}
	return 0, fmt.Errorf("no price for model %q: add its family to inputDollarsPerMillion", model)
}

type total struct {
	turns    int
	weighted float64
	dollars  float64
}

func (a *total) add(weighted, dollars float64) {
	a.turns++
	a.weighted += weighted
	a.dollars += dollars
}

func (t *transcript) print(w io.Writer) error {
	byModel, byCategory := map[string]*total{}, map[string]*total{}
	var all total
	for _, tn := range t.turns {
		perMillion, err := price(tn.model)
		if err != nil {
			return err
		}
		weighted := tn.weighted()
		dollars := weighted * perMillion / 1e6
		for _, group := range []struct {
			m   map[string]*total
			key string
		}{{byModel, tn.model}, {byCategory, tn.category}} {
			if group.m[group.key] == nil {
				group.m[group.key] = &total{}
			}
			group.m[group.key].add(weighted, dollars)
		}
		all.add(weighted, dollars)
	}
	wall := t.last.Sub(t.first).Truncate(time.Second)
	fmt.Fprintf(w, "turns %d\n", all.turns)
	fmt.Fprintf(w, "wall %dm%02ds\n", int(wall.Minutes()), int(wall.Seconds())%60)
	for _, name := range sortedBy(byModel, func(a, b string) bool { return a < b }) {
		fmt.Fprintf(w, "model %s turns=%d weighted=%.0f cost=$%.2f\n", name, byModel[name].turns, byModel[name].weighted, byModel[name].dollars)
	}
	for _, name := range sortedBy(byCategory, func(a, b string) bool {
		if byCategory[a].weighted != byCategory[b].weighted {
			return byCategory[a].weighted > byCategory[b].weighted
		}
		return a < b
	}) {
		fmt.Fprintf(w, "category %s turns=%d weighted=%.0f cost=$%.2f\n", name, byCategory[name].turns, byCategory[name].weighted, byCategory[name].dollars)
	}
	fmt.Fprintf(w, "total weighted=%.0f cost=$%.2f\n", all.weighted, all.dollars)
	return nil
}

func sortedBy(m map[string]*total, less func(a, b string) bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return less(keys[i], keys[j]) })
	return keys
}
