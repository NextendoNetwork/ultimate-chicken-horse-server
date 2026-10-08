package backend

import (
	"context"
	"testing"
	"time"
)

func TestMaintenanceExpiresIdleStateWithoutRequests(t *testing.T) {
	b, err := New(make([]byte, 32), []string{"test"}, &fakeRelay{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	b.sessions["expired"] = session{"old", time.Now().Add(-time.Second)}
	b.rooms["old"] = &room{expires: time.Now().Add(-time.Second)}
	b.profiles["old"] = &profile{last: time.Now().Add(-25 * time.Hour).UnixMilli()}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { b.Maintain(ctx); close(done) }()
	deadline := time.Now().Add(3 * time.Second)
	for {
		b.mu.Lock()
		empty := len(b.sessions) == 0 && len(b.rooms) == 0 && len(b.profiles) == 0
		b.mu.Unlock()
		if empty {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("idle state was not reclaimed")
		}
		time.Sleep(20 * time.Millisecond)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("maintenance did not stop")
	}
}
