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
		desc     string
		pushes   []terminalapi.Event
		wantPops []terminalapi.Event
	}{
		{
			desc: "empty queue returns nil",
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

	got, err := q.Pull(ctx)
	if err == nil {
		t.Fatal("Pull => expected timeout error, got nil")
	}

	close(ch)
	got, err = q.Pull(context.Background())
	if err != nil {
		t.Fatalf("Pull => unexpected error: %v", err)
	}

	if diff := pretty.Compare(want, got); diff != "" {
		t.Errorf("Pull => unexpected diff (-want, +got):\n%s", diff)
	}
}
