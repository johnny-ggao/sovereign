package config

import "testing"

func TestDSN(t *testing.T) {
	cfg := DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "test",
		Password: "pass",
		Name:     "testdb",
		SSLMode:  "disable",
	}

	dsn := cfg.DSN()
	expected := "host=localhost port=5432 user=test password=pass dbname=testdb sslmode=disable"
	if dsn != expected {
		t.Errorf("DSN() = %q, want %q", dsn, expected)
	}
}

func TestIntToStr(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{5432, "5432"},
		{1, "1"},
		{100, "100"},
	}

	for _, tt := range tests {
		got := intToStr(tt.input)
		if got != tt.want {
			t.Errorf("intToStr(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
