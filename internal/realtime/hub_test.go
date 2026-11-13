package realtime

import (
	"testing"
	"time"
)

func TestNotifyReachesOnlyOwner(t *testing.T) {
	h := NewHub()
	a, cancelA := h.Subscribe(1)
	defer cancelA()
	b, cancelB := h.Subscribe(2)
	defer cancelB()

	h.Notify(1)

	select {
	case <-a:
	case <-time.After(time.Second):
		t.Fatal("customer 1 tidak menerima sinyal")
	}
	select {
	case <-b:
		t.Fatal("customer 2 tidak boleh menerima sinyal customer 1")
	default:
	}
}

func TestNotifyCoalescesAndDoesNotBlock(t *testing.T) {
	h := NewHub()
	ch, cancel := h.Subscribe(1)
	defer cancel()

	for i := 0; i < 5; i++ {
		h.Notify(1) // tidak boleh macet walau tidak ada yang membaca
	}
	<-ch
	select {
	case <-ch:
		t.Fatal("sinyal beruntun seharusnya digabung jadi satu")
	default:
	}
}

func TestCancelUnsubscribesAndNilHubIsSafe(t *testing.T) {
	h := NewHub()
	_, cancel := h.Subscribe(1)
	cancel()
	if len(h.subs) != 0 {
		t.Fatalf("subs masih tersisa: %d", len(h.subs))
	}

	var nilHub *Hub
	nilHub.Notify(1)
	_, c := nilHub.Subscribe(1)
	c()
}

func TestNotifyStaffReachesAllStaffOnly(t *testing.T) {
	h := NewHub()
	s1, cancel1 := h.SubscribeStaff()
	defer cancel1()
	s2, cancel2 := h.SubscribeStaff()
	defer cancel2()
	c, cancelC := h.Subscribe(1)
	defer cancelC()

	h.NotifyStaff()

	for i, ch := range []<-chan struct{}{s1, s2} {
		select {
		case <-ch:
		case <-time.After(time.Second):
			t.Fatalf("kasir %d tidak menerima sinyal", i+1)
		}
	}
	select {
	case <-c:
		t.Fatal("customer tidak boleh menerima sinyal kasir")
	default:
	}
}

func TestStaffCancelUnsubscribesAndNilHubIsSafe(t *testing.T) {
	h := NewHub()
	_, cancel := h.SubscribeStaff()
	cancel()
	if len(h.staff) != 0 {
		t.Fatalf("staff masih tersisa: %d", len(h.staff))
	}

	var nilHub *Hub
	nilHub.NotifyStaff()
	_, c := nilHub.SubscribeStaff()
	c()
}
