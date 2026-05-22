package ticket

import (
	"testing"
	"time"
)

func TestStore_IssueConsume_HappyPath(t *testing.T) {
	s := NewStore(time.Minute)
	tk, err := s.Issue("sess-1", 42)
	if err != nil {
		t.Fatal(err)
	}
	uid, ok := s.Consume(tk, "sess-1")
	if !ok || uid != 42 {
		t.Fatalf("Consume() = (%d,%v); want (42,true)", uid, ok)
	}
	// 再消费一次必失败（single-use）。
	if _, ok := s.Consume(tk, "sess-1"); ok {
		t.Fatal("second Consume should fail")
	}
}

func TestStore_Consume_WrongSession(t *testing.T) {
	s := NewStore(time.Minute)
	tk, _ := s.Issue("sess-A", 1)
	if _, ok := s.Consume(tk, "sess-B"); ok {
		t.Fatal("Consume should reject mismatched session")
	}
	// 即使匹配回原 session 仍可消费？预期：不可以——上一次失败不应该删除条目，所以应可成功。
	if _, ok := s.Consume(tk, "sess-A"); !ok {
		t.Fatal("Consume with correct session should still succeed after a mismatched attempt")
	}
}

func TestStore_Consume_Expired(t *testing.T) {
	s := NewStore(10 * time.Millisecond)
	tk, _ := s.Issue("s", 1)
	time.Sleep(20 * time.Millisecond)
	if _, ok := s.Consume(tk, "s"); ok {
		t.Fatal("Consume should fail after TTL")
	}
}
