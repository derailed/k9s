// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package client

import (
	"log/slog"
	"os"
	"testing"
)

func TestMultiConnection_IsMultiContext(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name     string
		contexts []string
		want     bool
	}{
		{
			name:     "single context",
			contexts: []string{"context1"},
			want:     false,
		},
		{
			name:     "multiple contexts",
			contexts: []string{"context1", "context2"},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &MultiConnection{
				contexts:    tt.contexts,
				connections: make(map[string]Connection),
				configs:     make(map[string]*Config),
				log:         log,
			}

			if got := mc.IsMultiContext(); got != tt.want {
				t.Errorf("IsMultiContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultiConnection_ActiveContext(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tests := []struct {
		name     string
		contexts []string
		want     string
	}{
		{
			name:     "single context",
			contexts: []string{"my-context"},
			want:     "my-context",
		},
		{
			name:     "two contexts",
			contexts: []string{"context1", "context2"},
			want:     "multi:2",
		},
		{
			name:     "three contexts",
			contexts: []string{"context1", "context2", "context3"},
			want:     "multi:3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := &MultiConnection{
				contexts:    tt.contexts,
				primary:     tt.contexts[0],
				connections: make(map[string]Connection),
				configs:     make(map[string]*Config),
				log:         log,
			}

			if got := mc.ActiveContext(); got != tt.want {
				t.Errorf("ActiveContext() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMultiConnection_Contexts(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stdout, nil))

	contexts := []string{"context1", "context2", "context3"}
	mc := &MultiConnection{
		contexts:    contexts,
		connections: make(map[string]Connection),
		configs:     make(map[string]*Config),
		log:         log,
	}

	got := mc.Contexts()
	if len(got) != len(contexts) {
		t.Errorf("Contexts() length = %v, want %v", len(got), len(contexts))
	}

	for i, ctx := range got {
		if ctx != contexts[i] {
			t.Errorf("Contexts()[%d] = %v, want %v", i, ctx, contexts[i])
		}
	}

	// Ensure returned slice is a copy
	got[0] = "modified"
	if mc.contexts[0] == "modified" {
		t.Error("Contexts() should return a copy, not the original slice")
	}
}
