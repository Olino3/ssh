package system

import "testing"

func TestParseOSRelease(t *testing.T) {
	info := ParseOSRelease([]byte("PRETTY_NAME=\"Ubuntu 22.04.4 LTS\"\nID=ubuntu\nVERSION_ID=\"22.04\"\n"))
	if info.ID != "ubuntu" || info.Version != "22.04" || info.Pretty != "Ubuntu 22.04.4 LTS" {
		t.Fatalf("got %+v", info)
	}
}

func TestValidateOS(t *testing.T) {
	cases := []struct {
		id, ver string
		wantErr bool
	}{
		{"ubuntu", "22.04", false},
		{"ubuntu", "24.04", false},
		{"ubuntu", "19.10", true},
		{"debian", "11", false},
		{"debian", "12", false},
		{"debian", "10", true},
		{"arch", "", true},
		{"", "", true},
	}
	for _, c := range cases {
		err := ValidateOS(OSInfo{ID: c.id, Version: c.ver})
		if (err != nil) != c.wantErr {
			t.Errorf("%s %s: err=%v wantErr=%v", c.id, c.ver, err, c.wantErr)
		}
	}
}

func TestParseOSReleaseMissingKeysFailsClosed(t *testing.T) {
	info := ParseOSRelease([]byte("NAME=\"Ubuntu\"\n"))
	if info.ID != "" || info.Version != "" || info.Pretty != "" {
		t.Fatalf("expected zero-value OSInfo, got %+v", info)
	}
	if err := ValidateOS(info); err == nil {
		t.Fatal("expected ValidateOS to reject info with missing keys")
	}
}
