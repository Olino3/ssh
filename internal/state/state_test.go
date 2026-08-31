package state

import "testing"

func TestSaveLoadRoundtrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	s := &State{
		Role:     "server",
		Services: []string{"nas"},
		Keys:     []Key{{Name: "id_personal", Path: "/home/u/.ssh/id_personal", Comment: "u@h:personal"}},
	}
	s.MarkComplete("keys")
	if err := Save(s); err != nil {
		t.Fatal(err)
	}
	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.Role != "server" || !got.IsComplete("keys") || got.Keys[0].Name != "id_personal" {
		t.Fatalf("roundtrip mismatch: %+v", got)
	}
}

func TestLoadMissingReturnsNil(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	s, err := Load()
	if err != nil || s != nil {
		t.Fatalf("s=%v err=%v", s, err)
	}
}

func TestUpsertKeyAndHasService(t *testing.T) {
	s := &State{}
	s.UpsertKey(Key{Name: "id_personal"})
	s.UpsertKey(Key{Name: "id_personal", Comment: "x"})
	if len(s.Keys) != 1 || s.Keys[0].Comment != "x" {
		t.Fatalf("upsert failed: %+v", s.Keys)
	}
	s.AddService("nas")
	s.AddService("nas")
	if len(s.Services) != 1 {
		t.Fatalf("dup service: %+v", s.Services)
	}
	if !s.HasService("nas") {
		t.Fatal("HasService false")
	}
}
