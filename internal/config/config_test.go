package config

import "testing"

func TestRejectInvalidStartupConfiguration(t *testing.T) {
	for _, tc := range []struct{ key, value string }{
		{"REQUEST_TIMEOUT", "0s"}, {"CATALOG_TIMEOUT", "not-a-duration"},
		{"DB_MAX_OPEN", "0"}, {"DB_MAX_IDLE", "-1"}, {"DB_MAX_IDLE", "11"},
	} {
		t.Run(tc.key+"="+tc.value, func(t *testing.T) {
			t.Setenv("REQUEST_TIMEOUT", "2s")
			t.Setenv("CATALOG_TIMEOUT", "300ms")
			t.Setenv("DB_MAX_OPEN", "10")
			t.Setenv("DB_MAX_IDLE", "5")
			t.Setenv(tc.key, tc.value)
			if _, err := Load(); err == nil {
				t.Fatal("invalid configuration was accepted")
			}
		})
	}
}
