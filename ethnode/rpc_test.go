package ethnode

import "testing"

func TestParseUserAgent(t *testing.T) {
	testcases := []struct {
		clientVersion   string
		protocolVersion string
		netVersion      string
		wantKind        NodeKind
		wantNetwork     NetworkID
		wantFullNode    bool
	}{
		{"Geth/v1.8.16-unstable/linux-amd64/go1.10.3", "0x2712", "1", Geth, Mainnet, false},
		{"Geth/foo/v1.8.13-unstable/linux-amd64/go1.10.3", "0x3f", "1", Geth, Mainnet, true},
		{"Parity-Ethereum//v2.0.5-stable-7dc4d349a1-20180917/x86_64-linux-gnu/rustc1.29.0", "63", "1", Parity, Mainnet, true},
		{"Parity-Ethereum//v2.0.5-stable-7dc4d349a1-20180917/x86_64-linux-gnu/rustc1.29.0", "1", "1", Parity, Mainnet, false},
		{"pantheon/v1.1.3-dev-1d4946cd/linux-x86_64/oracle-java-11", "0x3f", "1", Pantheon, Mainnet, true},
	}

	for i, tc := range testcases {
		r, err := ParseUserAgent(tc.clientVersion, tc.protocolVersion, tc.netVersion)
		if err != nil {
			t.Errorf("[case %d] unexpected error for testcase: %s", i, err)
			continue
		}
		if r.Kind != tc.wantKind || r.Network != tc.wantNetwork || r.IsFullNode != tc.wantFullNode {
			t.Errorf("[case %d] wrong agent values: %+v", i, r)
		}
	}
}

func TestEnodeEqual(t *testing.T) {
	testcases := []struct {
		A, B string
		Want bool
	}{
		{"enode://foo@bar", "enode://foo@bar", true},
		{"enode://foo@bar", "enode://foo@bar", true},
		{"enode://foo@bar:30303", "enode://foo@[::]:30303", true},
		{"enode://foo@[::]:30303", "enode://foo@[::]:30303", true},
		{"enode://foo@[::]:30303", "enode://foo@1.1.1.1:30303", true},
		{"enode://foo@[::]:30303", "enode://foo@1.1.1.1:40404", false},
		{"enode://foo@[::]:30303", "enode://foo@", false},
		{"enode://foo@1.1.1.1:30303", "enode://foo@", false},
		{"enode://foo@1.1.1.1:30303", "enode://foo@2.2.2.2:30303", false},
		{"enode://foo@1.1.1.1:30303", "enode://foo@1.1.1.1:40404", false},
	}

	for i, tc := range testcases {
		got, want := EnodeEqual(tc.A, tc.B), tc.Want
		if got != want {
			t.Errorf("[case %d] %q ?= %q is %t", i, tc.A, tc.B, got)
		}
	}
}
