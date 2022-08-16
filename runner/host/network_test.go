package host

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/hcdiag/redact"
	"github.com/shirou/gopsutil/v3/net"
	"github.com/stretchr/testify/require"
)

func TestNetwork_convertInterfaceInfo(t *testing.T) {
	testCases := []struct {
		name               string
		network            Network
		inputInterfaceStat net.InterfaceStat
		expected           InterfaceInfo
		expectErr          bool
	}{
		{
			name:    "Test Interface",
			network: Network{},
			inputInterfaceStat: net.InterfaceStat{
				Index:        10,
				MTU:          1500,
				Name:         "eth0",
				HardwareAddr: "aa:bb:cc:dd:ee:ff",
				Flags: []string{
					"up",
					"loopback",
					"multicast",
				},
				Addrs: []net.InterfaceAddr{
					{
						Addr: "192.168.255.1/24",
					},
					{
						Addr: "fe80::1/64",
					},
				},
			},
			expected: InterfaceInfo{
				Index:        10,
				MTU:          1500,
				Name:         "eth0",
				HardwareAddr: "aa:bb:cc:dd:ee:ff",
				Flags: []string{
					"up",
					"loopback",
					"multicast",
				},
				Addrs: []InterfaceAddr{
					{
						Addr: "192.168.255.1/24",
					},
					{
						Addr: "fe80::1/64",
					},
				},
			},
		},
		{
			name: "Test Interface Redactions",
			network: Network{
				Redactions: createRedactionSlice(
					t,
					redact.Config{Matcher: "192.[\\d]{1,3}.[\\d]{1,3}.[\\d]{1,3}"},
					redact.Config{Matcher: "aa:bb:cc:dd:ee:ff"}),
			},
			inputInterfaceStat: net.InterfaceStat{
				Index:        10,
				MTU:          1500,
				Name:         "eth0",
				HardwareAddr: "aa:bb:cc:dd:ee:ff",
				Flags: []string{
					"up",
					"loopback",
					"multicast",
				},
				Addrs: []net.InterfaceAddr{
					{
						Addr: "192.168.255.1/24",
					},
					{
						Addr: "fe80::1/64",
					},
				},
			},
			expected: InterfaceInfo{
				Index:        10,
				MTU:          1500,
				Name:         "eth0",
				HardwareAddr: redact.DefaultReplace,
				Flags: []string{
					"up",
					"loopback",
					"multicast",
				},
				Addrs: []InterfaceAddr{
					{
						Addr: fmt.Sprintf("%s/24", redact.DefaultReplace),
					},
					{
						Addr: "fe80::1/64",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			interfaceInfo, err := tc.network.convertInterfaceInfo(tc.inputInterfaceStat)
			if tc.expectErr {
				require.Error(t, err, "an error was expected, but was not returned")
			} else {
				require.NoError(t, err)
				require.True(t, reflect.DeepEqual(tc.expected, interfaceInfo),
					"result did not match the expected result:\nactual=%#v\nexpected=%#v", interfaceInfo, tc.expected)
			}
		})
	}
}
