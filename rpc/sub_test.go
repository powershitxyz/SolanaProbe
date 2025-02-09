package rpc

import "testing"

func TestSubscribe(t *testing.T) {
	InitEssential()
}

func TestProcessSlot(t *testing.T) {
	slot := uint64(318831493)
	t.Logf("prepare for process slot: %d", slot)
	ProcessSlot(slot, nil)
}
