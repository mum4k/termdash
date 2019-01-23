// Copyright 2018 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package eventqueue

import (
	"context"
	"testing"
	"time"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/terminalapi"
)

func TestQueue(t *testing.T) {
	tests := []struct {
		desc      string
		pushes    []terminalapi.Event
		wantEmpty bool // Checked after pushes and before pops.
		wantPops  []terminalapi.Event
	}{
		{
			desc:      "empty queue returns nil",
			wantEmpty: true,
			wantPops: []terminalapi.Event{
				nil,
			},
		},
		{
			desc: "queue is FIFO",
			pushes: []terminalapi.Event{
				terminalapi.NewError("error1"),
				terminalapi.NewError("error2"),
				terminalapi.NewError("error3"),
			},
			wantEmpty: false,
			wantPops: []terminalapi.Event{
				terminalapi.NewError("error1"),
				terminalapi.NewError("error2"),
				terminalapi.NewError("error3"),
				nil,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			q := New()
			defer q.Close()
			for _, ev := range tc.pushes {
				q.Push(ev)
			}

			gotEmpty := q.Empty()
			if gotEmpty != tc.wantEmpty {
				t.Errorf("Empty => got %v, want %v", gotEmpty, tc.wantEmpty)
			}

			for i, want := range tc.wantPops {
				got := q.Pop()
				if diff := pretty.Compare(want, got); diff != "" {
					t.Errorf("Pop[%d] => unexpected diff (-want, +got):\n%s", i, diff)
				}
			}
		})
	}
}

func TestPullEventAvailable(t *testing.T) {
	q := New()
	defer q.Close()
	want := terminalapi.NewError("error event")
	q.Push(want)

	ctx := context.Background()
	got, err := q.Pull(ctx)
	if err != nil {
		t.Fatalf("Pull => unexpected error: %v", err)
	}
	if diff := pretty.Compare(want, got); diff != "" {
		t.Errorf("Pull => unexpected diff (-want, +got):\n%s", diff)
	}
}

func TestPullBlocksUntilAvailable(t *testing.T) {
	q := New()
	defer q.Close()
	want := terminalapi.NewError("error event")

	ch := make(chan struct{})
	go func() {
		<-ch
		q.Push(want)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	if _, err := q.Pull(ctx); err != nil {
		t.Fatal("Pull => expected timeout error, got nil")
	}

	close(ch)
	got, err := q.Pull(context.Background())
	if err != nil {
		t.Fatalf("Pull => unexpected error: %v", err)
	}

	if diff := pretty.Compare(want, got); diff != "" {
		t.Errorf("Pull => unexpected diff (-want, +got):\n%s", diff)
	}
}
