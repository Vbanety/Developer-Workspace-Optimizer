package linux

import (
	"os/exec"
	"testing"
)

func TestParseSnapListDisabled(t *testing.T) {
	dir := t.TempDir()
	orig := snapsDir
	snapsDir = dir
	t.Cleanup(func() { snapsDir = orig })

	writeFixtureFile(t, dir, "cups_1206.snap", 100)
	writeFixtureFile(t, dir, "cups_1225.snap", 200)
	writeFixtureFile(t, dir, "firefox_6316.snap", 300)
	writeFixtureFile(t, dir, "android-studio_232.snap", 400)

	output := `Name                Version                         Rev    Tracking         Publisher         Notes
cups                2.4.19-2                        1206   latest/stable    openprinting**    disabled
cups                2.4.19-2                        1225   latest/stable    openprinting**    -
firefox             139.0.4-1                       6316   latest/stable    mozilla**         disabled
android-studio      2026.1.1.10-quail1-patch2       232    latest/stable    snapcrafters*     classic,disabled
`
	revisions, err := parseSnapListDisabled(output)
	if err != nil {
		t.Fatal(err)
	}
	if len(revisions) != 3 {
		t.Fatalf("expected 3 disabled revisions, got %d: %+v", len(revisions), revisions)
	}

	var total int64
	for _, r := range revisions {
		total += r.sizeBytes
	}
	if total != 100+300+400 {
		t.Fatalf("expected total 800, got %d", total)
	}
}

func TestParseSnapListDisabledHeaderOnly(t *testing.T) {
	revisions, err := parseSnapListDisabled("Name    Version    Rev    Tracking    Publisher    Notes\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(revisions) != 0 {
		t.Fatalf("expected no revisions from header-only output, got %d", len(revisions))
	}
}

func TestParseSnapListDisabledSkipsMalformedLines(t *testing.T) {
	output := "Name Version Rev Tracking Publisher Notes\nnot-enough-fields\n"
	revisions, err := parseSnapListDisabled(output)
	if err != nil {
		t.Fatal(err)
	}
	if len(revisions) != 0 {
		t.Fatalf("expected malformed line to be skipped, got %d revisions", len(revisions))
	}
}

func TestSnapDetectAndCalculate(t *testing.T) {
	if _, err := exec.LookPath("snap"); err != nil {
		t.Skip("snap not installed, skipping integration check")
	}

	s := NewSnap()
	found, err := s.Detect()
	if err != nil {
		t.Fatal(err)
	}
	if !found {
		t.Skip("snap CLI present but not usable here, skipping")
	}

	finding, err := s.Calculate()
	if err != nil {
		t.Fatal(err)
	}
	if finding.SizeBytes < 0 {
		t.Fatalf("expected non-negative SizeBytes, got %d", finding.SizeBytes)
	}
}
