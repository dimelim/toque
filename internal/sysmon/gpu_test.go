package sysmon

import (
	"testing"
	"time"
)

func TestGPUByProcessTakesBusiestEngineType(t *testing.T) {
	got := gpuByProcess(map[string]float64{
		"pid_10_luid_0x0_0xD1C5_phys_0_eng_0_engtype_3D":          30,
		"pid_10_luid_0x0_0xD1C5_phys_0_eng_1_engtype_3D":          5,
		"pid_10_luid_0x0_0xD1C5_phys_0_eng_3_engtype_VideoDecode": 20,
		"pid_20_luid_0x0_0xD1C5_phys_0_eng_0_engtype_Copy":        4,
		"garbage": 99,
	})
	if got[10] != 35 || got[20] != 4 || len(got) != 2 {
		t.Fatalf("gpuByProcess = %v", got)
	}
}

func TestGPUSamplerReadsThisMachine(t *testing.T) {
	g, err := NewGPUSampler()
	if err != nil {
		t.Skip(err)
	}
	defer g.Close()
	time.Sleep(time.Second)
	t.Logf("%d procesos con GPU", len(g.Sample()))
}
