package roblox

import "testing"

func TestParseLineReadsDatacenter(t *testing.T) {
	line := "2026-10-27T15:21:52.973Z,2.973706,4a04,6,Info [DFLog::NetworkClient] Transport selection: useRbxTransportEnabled=false, selectedTransport=RakNet, rccPort=58698, DatacenterId=25774"
	ev, ok := ParseLine(line)
	if !ok || ev.Kind != EventDatacenter || ev.Datacenter != 25774 {
		t.Fatalf("ParseLine = %+v, %v; want datacenter 25774", ev, ok)
	}
}
