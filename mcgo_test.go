package mcgo

import "testing"

func TestVersions(t *testing.T) {
	v, err := Versions()
	if err != nil {
		t.Fatalf("Versions: %v", err)
	}
	if len(v) == 0 {
		t.Fatal("no versions")
	}
	t.Logf("got %d versions", len(v))
}

func TestLatest(t *testing.T) {
	rel, err := Latest("release")
	if err != nil {
		t.Fatalf("Latest(release): %v", err)
	}
	if rel == "" {
		t.Fatal("empty release")
	}
	t.Logf("latest release: %s", rel)
}

func TestOfflineAuth(t *testing.T) {
	p, err := OfflineAuth{"Notch"}.Login()
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if p.Username != "Notch" {
		t.Errorf("username = %q", p.Username)
	}
	if p.UUID == "" {
		t.Error("uuid empty")
	}
}

func TestFabricLoader(t *testing.T) {
	v, err := LoaderVersion(Fabric, "1.21.4")
	if err != nil {
		t.Fatalf("LoaderVersion: %v", err)
	}
	t.Logf("fabric loader for 1.21.4: %s", v)
}

func TestLoaders(t *testing.T) {
	loaders := Loaders()
	if len(loaders) != 5 {
		t.Errorf("expected 5 loaders, got %d", len(loaders))
	}
}

func TestEventBus(t *testing.T) {
	bus := NewEventBus()
	got := false
	bus.On(func(e Event) { got = true })
	bus.Emit(Event{Type: EvtInstallCompleted})
	if !got {
		t.Error("event not received")
	}
}
