package winx

import (
	"os"
	"testing"

	"golang.org/x/sys/windows"
)

func TestEcoRoundTripOnOwnProcess(t *testing.T) {
	pid := uint32(os.Getpid())
	before, err := EcoOf(pid)
	if err != nil {
		t.Fatal(err)
	}
	if err := SetEco(pid, Efficient()); err != nil {
		t.Fatal(err)
	}
	during, err := EcoOf(pid)
	if err != nil {
		t.Fatal(err)
	}
	if during.Class != windows.IDLE_PRIORITY_CLASS {
		t.Fatalf("class during = %#x", during.Class)
	}
	if during.Known && during.State&throttleExecutionSpeed == 0 {
		t.Fatalf("eco not set: %+v", during)
	}
	if err := SetEco(pid, before.Restorable()); err != nil {
		t.Fatal(err)
	}
	after, _ := EcoOf(pid)
	if after.Class != before.Class || after.Control != before.Control || after.State != before.State {
		t.Fatalf("not restored: before %+v after %+v", before, after)
	}
	t.Logf("before %+v during %+v after %+v", before, during, after)
}

func TestAudiblePIDsRuns(t *testing.T) {
	pids, err := AudiblePIDs()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("audible: %v", pids)
}
