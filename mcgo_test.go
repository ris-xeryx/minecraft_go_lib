package mcgo

import "testing"

func TestGetAllVersions(t *testing.T) {
	versions, err := GetAllVersions()
	if err != nil {
		t.Fatalf("GetAllVersions failed: %v", err)
	}
	if len(versions) == 0 {
		t.Fatal("no versions returned")
	}
	t.Logf("Got %d versions", len(versions))
}

func TestOfflineUUID(t *testing.T) {
	p := NewOfflineAuth("Notch")
	prof, err := p.Authenticate()
	if err != nil {
		t.Fatalf("Authenticate failed: %v", err)
	}
	if prof.Username != "Notch" {
		t.Errorf("username = %q, want Notch", prof.Username)
	}
	if prof.UUID == "" {
		t.Error("UUID is empty")
	}
}

func TestFabricLoader(t *testing.T) {
	versions, err := GetLoaderVersions(LoaderFabric, "1.21.4")
	if err != nil {
		t.Fatalf("Fabric loader failed: %v", err)
	}
	if len(versions) == 0 {
		t.Skip("no fabric versions found (network?)")
	}
	t.Logf("Fabric loader: %s", versions[0])
}
