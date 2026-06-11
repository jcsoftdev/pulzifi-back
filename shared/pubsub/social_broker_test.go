package pubsub

import (
	"testing"
	"time"
)

func TestSocialBroker_PublishDeliversToSubscriber(t *testing.T) {
	b := NewSocialBroker()
	ch, unsub := b.Subscribe("profile-1")
	defer unsub()

	b.Publish("profile-1", []byte("hello"))

	select {
	case got := <-ch:
		if string(got) != "hello" {
			t.Errorf("expected hello, got %s", got)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for payload")
	}
}

func TestSocialBroker_ReplaysCacheToLateSubscriber(t *testing.T) {
	b := NewSocialBroker()
	b.Publish("profile-2", []byte("cached"))

	ch, unsub := b.Subscribe("profile-2")
	defer unsub()

	select {
	case got := <-ch:
		if string(got) != "cached" {
			t.Errorf("expected cached, got %s", got)
		}
	case <-time.After(time.Second):
		t.Fatal("expected cached replay")
	}
}

func TestSocialBroker_UnsubscribeRemovesListener(t *testing.T) {
	b := NewSocialBroker()
	ch, unsub := b.Subscribe("profile-3")
	unsub()

	_, open := <-ch
	if open {
		t.Error("expected channel closed after unsubscribe")
	}
}
