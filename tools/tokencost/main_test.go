package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_PrintsTheCostOfTheFixtureTranscript(t *testing.T) {
	want, err := os.ReadFile("testdata/expected.txt")
	require.NoError(t, err)
	var out bytes.Buffer

	err = run([]string{"testdata/transcript.jsonl"}, strings.NewReader(""), &out)

	require.NoError(t, err)
	assert.Equal(t, string(want), out.String())
}

func TestRun_ReadsStdin_WhenNoFileIsGiven(t *testing.T) {
	raw, err := os.ReadFile("testdata/transcript.jsonl")
	require.NoError(t, err)
	want, err := os.ReadFile("testdata/expected.txt")
	require.NoError(t, err)
	var out bytes.Buffer

	err = run(nil, bytes.NewReader(raw), &out)

	require.NoError(t, err)
	assert.Equal(t, string(want), out.String())
}

func TestRun_Fails(t *testing.T) {
	tests := []struct {
		name, input, wantErr string
	}{
		{"on an empty transcript", "", "no assistant message with usage"},
		{"on a line that is not JSON", "{\n", "line 1"},
		{"on a model it has no price for", `{"type":"assistant","message":{"id":"m","model":"gpt-9","usage":{"input_tokens":1}}}` + "\n", `no price for model "gpt-9"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := run(nil, strings.NewReader(tt.input), &bytes.Buffer{})

			assert.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestRun_CountsAMessageOnce_WhateverItsNumberOfContentBlocks(t *testing.T) {
	block := `{"type":"assistant","timestamp":"2026-11-04T10:00:00Z","message":{"id":"m","model":"claude-haiku-5","content":[{"type":"text"}],"usage":{"output_tokens":1000000}}}` + "\n"
	var out bytes.Buffer

	err := run(nil, strings.NewReader(block+block+block), &out)

	require.NoError(t, err)
	assert.Equal(t, "turns 1\nwall 0m00s\nmodel claude-haiku-5 turns=1 weighted=5000000 cost=$5.00\ncategory text turns=1 weighted=5000000 cost=$5.00\ntotal weighted=5000000 cost=$5.00\n", out.String())
}

func TestRun_FailsOnTwoArguments(t *testing.T) {
	err := run([]string{"a", "b"}, strings.NewReader(""), &bytes.Buffer{})

	assert.ErrorContains(t, err, "usage: tokencost")
}
