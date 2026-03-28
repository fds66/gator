package main

import (
	"os/exec"
	"strings"
	"testing"
)

/*
	func TestHandleFollow(t *testing.T) {
		// set up a test state with a mock or real test DB
		s := &State{  }
		cmd := Command{
			Name:      "follow",
			Arguments: []string{"https://hnrss.org/newest"},
		}
		err := handlerFollow(s, cmd)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	}
*/
// Set of tests needed to be run in order because they use the database
// testing registering users and adding feeds
func TestReset(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "reset")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\noutput: %s", err, out)
	}

}
func TestRegister(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "register", "kahya")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\noutput: %s", err, out)
	}
	if !strings.Contains(string(out), "kahya") {
		t.Errorf("expected user name in output, got: %s", out)
	}
}
func TestAddFeed(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "addfeed", "Hacker News RSS", "https://hnrss.org/newest")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\noutput: %s", err, out)
	}
	if !strings.Contains(string(out), "Hacker News") {
		t.Errorf("expected feed name in output, got: %s", out)
	}
}
func TestRegister2(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "register", "holgith")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\noutput: %s", err, out)
	}
	if !strings.Contains(string(out), "holgith") {
		t.Errorf("expected user name in output, got: %s", out)
	}
}
func TestAddFeed2(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "addfeed", "Lanes Blog", "https://www.wagslane.dev/index.xml")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\noutput: %s", err, out)
	}
	if !strings.Contains(string(out), "Lanes Blog") {
		t.Errorf("expected feed name in output, got: %s", out)
	}
}
func TestRegister3(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "register", "ballan")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\noutput: %s", err, out)
	}
	if !strings.Contains(string(out), "ballan") {
		t.Errorf("expected user name in output, got: %s", out)
	}
}
func TestFeeds(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "feeds")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\noutput: %s", err, out)
	}
	if !strings.Contains(string(out), "Hacker News RSS") {
		t.Errorf("expected feed name in output, got: %s", out)
	}
	if !strings.Contains(string(out), "kahya") {
		t.Errorf("expected user name in output, got: %s", out)
	}
	if !strings.Contains(string(out), "Lanes Blog") {
		t.Errorf("expected feed name in output, got: %s", out)
	}
	if !strings.Contains(string(out), "holgith") {
		t.Errorf("expected user names in output, got: %s", out)
	}
	if strings.Contains(string(out), "ballan") {
		t.Errorf("expected user names in output to not include ballan, got: %s", out)
	}

}

// testing follows

func TestFollowCommand(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "follow", "https://hnrss.org/newest")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("command failed: %v\noutput: %s", err, out)
	}
	if !strings.Contains(string(out), "Hacker News RSS") {
		t.Errorf("expected feed name in output, got: %s", out)
	}
}

/*
Test expecting to exit with 1
func TestFollowRequiresLogin(t *testing.T) {
    s := &state{  no logged-in user  }
    cmd := command{
        name: "follow",
        args: []string{"https://hnrss.org/newest"},
    }
    err := handlerFollow(s, cmd)
    if err == nil {
        t.Error("expected an error when no user is logged in, got nil")
    }
}


func TestFollowWithNoArgs(t *testing.T) {
    cmd := exec.Command("go", "run", ".", "follow")
    err := cmd.Run()

    var exitErr *exec.ExitError
    if !errors.As(err, &exitErr) {
        t.Fatal("expected a non-zero exit code")
    }
    if exitErr.ExitCode() != 1 {
        t.Errorf("expected exit code 1, got %d", exitErr.ExitCode())
    }
}

func runGator(t *testing.T, args ...string) (string, int) {
    t.Helper()
    cmd := exec.Command("go", append([]string{"run", "."}, args...)...)
    out, err := cmd.CombinedOutput()
    if err != nil {
        var exitErr *exec.ExitError
        if errors.As(err, &exitErr) {
            return string(out), exitErr.ExitCode()
        }
        t.Fatalf("unexpected error: %v", err)
    }
    return string(out), 0
}

// Usage:
func TestFollowWithNoArgs(t *testing.T) {
    _, code := runGator(t, "follow")
    if code != 1 {
        t.Errorf("expected exit code 1, got %d", code)
    }
}
*/
