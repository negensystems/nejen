package main

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestParseCleanDuration(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
	}{
		{"30", 30 * time.Second},
		{"45s", 45 * time.Second},
		{"2m", 2 * time.Minute},
		{"1m30s", 90 * time.Second},
		{"  20  ", 20 * time.Second},
		// Out of range clamps rather than failing: a typo should still
		// clean the keyboard, just not for ten hours.
		{"10h", cleanMax},
		{"1", cleanMin},
		{"0", cleanMin},
		{"-5", cleanMin},
	}
	for _, c := range cases {
		got, err := parseCleanDuration(c.in)
		if err != nil {
			t.Errorf("parseCleanDuration(%q) returned error: %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("parseCleanDuration(%q) = %v, want %v", c.in, got, c.want)
		}
	}

	for _, bad := range []string{"", "   ", "soon", "5 minutes", "m"} {
		if _, err := parseCleanDuration(bad); err == nil {
			t.Errorf("parseCleanDuration(%q) accepted a duration it should not have", bad)
		}
	}
}

func TestHumanCleanDuration(t *testing.T) {
	cases := []struct {
		in   time.Duration
		want string
	}{
		{30 * time.Second, "30 seconds"},
		{59 * time.Second, "59 seconds"},
		{time.Minute, "1 minute"},
		{2 * time.Minute, "2 minutes"},
		{90 * time.Second, "1m30s"},
		{125 * time.Second, "2m05s"},
	}
	for _, c := range cases {
		if got := humanCleanDuration(c.in); got != c.want {
			t.Errorf("humanCleanDuration(%v) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestCleanStateRoundTrip(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())

	if _, _, ok := readCleanState(); ok {
		t.Fatal("readCleanState reported a session with no state file")
	}

	deadline := time.Now().Add(cleanDefault).Truncate(time.Second)
	if err := writeCleanState(4242, deadline); err != nil {
		t.Fatalf("writeCleanState: %v", err)
	}
	pid, got, ok := readCleanState()
	if !ok {
		t.Fatal("readCleanState did not read back the state it just wrote")
	}
	if pid != 4242 {
		t.Errorf("pid = %d, want 4242", pid)
	}
	if !got.Equal(deadline) {
		t.Errorf("deadline = %v, want %v", got, deadline)
	}

	// Garbage in the state file must read as "no session", never as a
	// session with a pid of zero -- stopClean signals whatever it reads.
	for _, junk := range []string{"", "nonsense", "1", "abc 123", "123 abc", "1 2 3"} {
		if err := os.WriteFile(cleanStatePath(), []byte(junk), 0644); err != nil {
			t.Fatal(err)
		}
		if _, _, ok := readCleanState(); ok {
			t.Errorf("readCleanState accepted %q as a session", junk)
		}
	}
}

func TestCleanLiveRejectsStaleSessions(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", t.TempDir())

	// Our own pid is certainly alive, so these cases isolate the deadline.
	if err := writeCleanState(os.Getpid(), time.Now().Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if _, _, live := cleanLive(); !live {
		t.Error("cleanLive said a running, unexpired session was not live")
	}

	if err := writeCleanState(os.Getpid(), time.Now().Add(-time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, _, live := cleanLive(); live {
		t.Error("cleanLive said an expired session was still live")
	}

	// A pid that cannot exist stands in for a supervisor that was killed.
	if err := writeCleanState(0, time.Now().Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if _, _, live := cleanLive(); live {
		t.Error("cleanLive said a session with a dead supervisor was live")
	}
}

func TestCleanOverlayLinesCountdown(t *testing.T) {
	cases := []struct {
		in   time.Duration
		want string
	}{
		{30 * time.Second, "0:30"},
		{9 * time.Second, "0:09"},
		{90 * time.Second, "1:30"},
		{5 * time.Minute, "5:00"},
		// The overlay redraws faster than once a second, so it sees
		// fractional and (briefly) negative remainders.
		{1500 * time.Millisecond, "0:02"},
		{-3 * time.Second, "0:00"},
	}
	for _, c := range cases {
		lines := cleanOverlayLines(c.in)
		found := false
		for _, l := range lines {
			if strings.TrimSpace(l) == c.want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("cleanOverlayLines(%v) has no line %q, got %q", c.in, c.want, lines)
		}
	}
}
