package tweaks

import "testing"

func TestDirectXSettingKeepsOtherEntries(t *testing.T) {
	cases := map[string]string{
		"":                     "SwapEffectUpgradeEnable=1;",
		"VRROptimizeEnable=0;": "VRROptimizeEnable=0;SwapEffectUpgradeEnable=1;",
		"SwapEffectUpgradeEnable=0;AutoHDREnable=1;": "SwapEffectUpgradeEnable=1;AutoHDREnable=1;",
		"AutoHDREnable=1": "AutoHDREnable=1;SwapEffectUpgradeEnable=1;",
	}
	for in, want := range cases {
		if got := withDirectXSetting(in, swapUpgrade, "1"); got != want {
			t.Errorf("withDirectXSetting(%q) = %q, want %q", in, got, want)
		}
		if got := directXSetting(want, swapUpgrade); got != "1" {
			t.Errorf("directXSetting(%q) = %q", want, got)
		}
	}
}

func TestDirectXSettingRemovesOnlyItsEntry(t *testing.T) {
	cases := map[string]string{
		"SwapEffectUpgradeEnable=1;":                     "",
		"VRROptimizeEnable=1;SwapEffectUpgradeEnable=1;": "VRROptimizeEnable=1;",
		"AutoHDREnable=1;":                               "AutoHDREnable=1;",
	}
	for in, want := range cases {
		if got := withDirectXSetting(in, swapUpgrade, ""); got != want {
			t.Errorf("withDirectXSetting(%q, \"\") = %q, want %q", in, got, want)
		}
	}
}
