package main

import "testing"

func TestInterfaceTraffic(t *testing.T) {
	routes := `Iface Destination Gateway Flags RefCnt Use Metric Mask MTU Window IRTT
tun0 01000000 00000000 0001 0 0 0 01000000 0 0 0
eth1 00000000 0100000A 0003 0 0 50 00000000 0 0 0
eth0 00000000 010000AC 0003 0 0 10 00000000 0 0 0`
	devices := `tun0: 100 0 0 0 0 0 0 0 200 0 0 0 0 0 0 0
 eth0: 15000 0 0 0 0 0 0 0 732 0 0 0 0 0 0 0
 eth1: 900 0 0 0 0 0 0 0 2000 0 0 0 0 0 0 0`
	want := traffic{Interface: "eth0", Received: 15000, Sent: 732, SampledAt: 1234}
	if got := interfaceTraffic(routes, devices, 1234); got == nil || *got != want {
		t.Fatalf("interfaceTraffic() = %+v, want %+v", got, want)
	}
	if got := interfaceTraffic(routes, "eth0: 0", 1234); got != nil {
		t.Fatalf("interfaceTraffic() with truncated counters = %+v, want nil", got)
	}
}
