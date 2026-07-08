package linux

import (
	"os/exec"
	"testing"
)

func TestParseDockerSize(t *testing.T) {
	cases := []struct {
		in   string
		want int64
	}{
		{"0B", 0},
		{"0B (0%)", 0},
		{"14.5kB", 14500},
		{"1.37GB (70%)", 1370000000},
		{"1.955GB", 1955000000},
		{"2TB", 2_000_000_000_000},
	}
	for _, c := range cases {
		got, err := parseDockerSize(c.in)
		if err != nil {
			t.Fatalf("parseDockerSize(%q): %v", c.in, err)
		}
		if got != c.want {
			t.Fatalf("parseDockerSize(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestParseDockerSizeRejectsUnknownUnit(t *testing.T) {
	if _, err := parseDockerSize("1.2XB"); err == nil {
		t.Fatal("expected an error for an unrecognized unit")
	}
}

func TestDockerDetectAndCalculate(t *testing.T) {
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not installed, skipping integration check")
	}

	d := NewDocker()
	found, err := d.Detect()
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Skip("docker CLI present but daemon unreachable, skipping")
	}

	finding, err := d.Calculate()
	if err != nil {
		t.Fatal(err)
	}
	if finding.SizeBytes < 0 {
		t.Fatalf("expected non-negative SizeBytes, got %d", finding.SizeBytes)
	}
}
