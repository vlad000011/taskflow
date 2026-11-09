package service

import "testing"

func TestValidTaskStatus(t *testing.T) {
	for _, s := range []string{"backlog", "todo", "in_progress", "review", "done"} {
		if !ValidTaskStatus(s) {
			t.Fatalf("expected status %q to be valid", s)
		}
	}
	if ValidTaskStatus("finished") {
		t.Fatal("unknown status must be rejected")
	}
}
func TestValidPriority(t *testing.T) {
	if !ValidPriority("high") || ValidPriority("urgent-ish") {
		t.Fatal("priority validation failed")
	}
}

func TestPasswordHashing(t *testing.T) {
	password := "a-strong-demo-password"
	hash := HashPassword(password)
	if hash == password || !CheckPassword(hash, password) || CheckPassword(hash, "wrong-password") {
		t.Fatal("password hashing or verification failed")
	}
}
