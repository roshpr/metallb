// SPDX-License-Identifier:Apache-2.0

package config

import (
	"strings"
	"testing"
	"time"

	"go.universe.tf/metallb/api/v1beta1"
	"go.universe.tf/metallb/api/v1beta2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		desc     string
		config   ClusterResources
		mustFail bool
	}{
		{
			desc: "peer with bfd profile",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					{
						Spec: v1beta2.BGPPeerSpec{
							Address:    "1.2.3.4",
							BFDProfile: "foo",
						},
					},
				},
			},
			mustFail: true,
		},
		{
			desc: "bfd profile set",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					{
						Spec: v1beta2.BGPPeerSpec{
							Address: "1.2.3.4",
						},
					},
				},
				BFDProfiles: []v1beta1.BFDProfile{
					{
						ObjectMeta: v1.ObjectMeta{Name: "foo"},
					},
				},
			},
			mustFail: true,
		},
		{
			desc: "v6 address",
			config: ClusterResources{
				Pools: []v1beta1.IPAddressPool{
					{
						Spec: v1beta1.IPAddressPoolSpec{
							Addresses: []string{"2001:db8::/64"},
						},
					},
				},
				BGPAdvs: []v1beta1.BGPAdvertisement{
					{
						ObjectMeta: v1.ObjectMeta{
							Name: "foo",
						},
					},
				},
			},
			mustFail: true,
		},
		{
			desc: "v6 address but pool not selected",
			config: ClusterResources{
				Pools: []v1beta1.IPAddressPool{
					{
						ObjectMeta: v1.ObjectMeta{
							Name: "foo",
						},
						Spec: v1beta1.IPAddressPoolSpec{
							Addresses: []string{"2001:db8::/64"},
						},
					},
				},
				BGPAdvs: []v1beta1.BGPAdvertisement{
					{
						ObjectMeta: v1.ObjectMeta{
							Name: "foo",
						},
						Spec: v1beta1.BGPAdvertisementSpec{
							IPAddressPools: []string{"bar"},
						},
					},
				},
			},
		},
		{
			desc: "v6 address and selected",
			config: ClusterResources{
				Pools: []v1beta1.IPAddressPool{
					{
						ObjectMeta: v1.ObjectMeta{
							Name: "foo",
						},
						Spec: v1beta1.IPAddressPoolSpec{
							Addresses: []string{"2001:db8::/64"},
						},
					},
				},
				BGPAdvs: []v1beta1.BGPAdvertisement{
					{
						ObjectMeta: v1.ObjectMeta{
							Name: "bar",
						},
						Spec: v1beta1.BGPAdvertisementSpec{
							IPAddressPools: []string{"foo1"},
						},
					},
					{
						ObjectMeta: v1.ObjectMeta{
							Name: "bar",
						},
						Spec: v1beta1.BGPAdvertisementSpec{
							IPAddressPools: []string{"foo"},
						},
					},
				},
			},
			mustFail: true,
		},
		{
			desc: "v6 address and selected by labels",
			config: ClusterResources{
				Pools: []v1beta1.IPAddressPool{
					{
						ObjectMeta: v1.ObjectMeta{
							Name:   "foo",
							Labels: map[string]string{"key": "value"},
						},
						Spec: v1beta1.IPAddressPoolSpec{
							Addresses: []string{"2001:db8::/64"},
						},
					},
				},
				BGPAdvs: []v1beta1.BGPAdvertisement{
					{
						ObjectMeta: v1.ObjectMeta{
							Name: "bar",
						},
						Spec: v1beta1.BGPAdvertisementSpec{
							IPAddressPools: []string{"foo1"},
						},
					},
					{
						ObjectMeta: v1.ObjectMeta{
							Name: "bar",
						},
						Spec: v1beta1.BGPAdvertisementSpec{
							IPAddressPoolSelectors: []v1.LabelSelector{
								{
									MatchLabels: map[string]string{"key": "value"},
								},
							},
						},
					},
				},
			},
			mustFail: true,
		},
		{
			desc: "enable BGP GracefulRestart",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					{
						Spec: v1beta2.BGPPeerSpec{
							EnableGracefulRestart: true,
						},
					},
				},
			},
			mustFail: true,
		},
		{
			desc: "disable BGP MP",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					{
						Spec: v1beta2.BGPPeerSpec{
							DisableMP: true, //nolint:staticcheck // kept for backwards compatibility
						},
					},
				},
			},
			mustFail: true,
		},
		{
			desc: "dynamic ASN",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					{
						Spec: v1beta2.BGPPeerSpec{
							DynamicASN: "internal",
						},
					},
				},
			},
			mustFail: true,
		},
		{
			desc: "interface",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					{
						Spec: v1beta2.BGPPeerSpec{
							Interface: "eth0",
						},
					},
				},
			},
			mustFail: true,
		},
		{
			desc: "localASN",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					{
						Spec: v1beta2.BGPPeerSpec{
							LocalASN: 65410,
						},
					},
				},
			},
			mustFail: true,
		},
		{
			desc: "keepalive time",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					{
						Spec: v1beta2.BGPPeerSpec{
							KeepaliveTime: &v1.Duration{Duration: time.Second},
						},
					},
				},
			},
			mustFail: true,
		},
		{
			desc: "connect time",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					{
						Spec: v1beta2.BGPPeerSpec{
							ConnectTime: ptr.To(v1.Duration{Duration: time.Second}),
						},
					},
				},
			},
			mustFail: true,
		},
		{
			desc: "large BGP community inside BGP Advertisement",
			config: ClusterResources{
				BGPAdvs: []v1beta1.BGPAdvertisement{
					{
						Spec: v1beta1.BGPAdvertisementSpec{
							Communities: []string{"large:123:456:789"},
						},
					},
				},
			},
			mustFail: true,
		},
		{
			desc: "large BGP community inside Community CR",
			config: ClusterResources{
				Communities: []v1beta1.Community{
					{
						Spec: v1beta1.CommunitySpec{
							Communities: []v1beta1.CommunityAlias{
								{
									Value: "large:123:456:789",
								},
							},
						},
					},
				},
			},
			mustFail: true,
		},
		{
			desc: "should pass",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					{
						Spec: v1beta2.BGPPeerSpec{
							Address: "1.2.3.4",
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			err := DiscardFRROnly(test.config)
			if test.mustFail && err == nil {
				t.Fatalf("Expected error for %s", test.desc)
			}
			if !test.mustFail && err != nil {
				t.Fatalf("Not expected error %s for %s", err, test.desc)
			}
		})
	}
}

func TestValidateFRR(t *testing.T) {
	tests := []struct {
		desc     string
		config   ClusterResources
		mustFail bool
	}{
		{
			desc: "peer with routerid",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					{
						Spec: v1beta2.BGPPeerSpec{
							Address:  "1.2.3.4",
							RouterID: "1.2.3.4",
						},
					},
				},
			},
		},
		{
			desc: "routerid set, one different",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					{
						Spec: v1beta2.BGPPeerSpec{
							Address:  "1.2.3.4",
							RouterID: "1.2.3.4",
						},
					},
					{
						Spec: v1beta2.BGPPeerSpec{
							Address:  "1.2.3.5",
							RouterID: "1.2.3.4",
						},
					},
					{
						Spec: v1beta2.BGPPeerSpec{
							Address:  "1.2.3.6",
							RouterID: "1.2.3.5",
						},
					},
				},
			},
			mustFail: true,
		},
		{
			desc: "routerid set, one not set",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					{
						Spec: v1beta2.BGPPeerSpec{
							Address:  "1.2.3.4",
							RouterID: "1.2.3.4",
						},
					},
					{
						Spec: v1beta2.BGPPeerSpec{
							Address:  "1.2.3.5",
							RouterID: "1.2.3.4",
						},
					},
					{
						Spec: v1beta2.BGPPeerSpec{
							Address: "1.2.3.6",
						},
					},
				},
			},
			mustFail: true,
		},
		{
			desc: "bfd profile set",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					{
						Spec: v1beta2.BGPPeerSpec{
							Address:  "1.2.3.4",
							RouterID: "1.2.3.4",
						},
					},
				},
				BFDProfiles: []v1beta1.BFDProfile{
					{
						ObjectMeta: v1.ObjectMeta{
							Name: "foo",
						},
					},
				},
			},
		},
		{
			desc: "myAsn set, all equals",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					{
						Spec: v1beta2.BGPPeerSpec{
							Address: "1.2.3.4",
							MyASN:   123,
						},
					},
					{
						Spec: v1beta2.BGPPeerSpec{
							Address: "1.2.3.5",
							MyASN:   123,
						},
					},
					{
						Spec: v1beta2.BGPPeerSpec{
							Address: "1.2.3.6",
							MyASN:   123,
						},
					},
				},
			},
		},
		{
			desc: "myAsn set, one different",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					{
						Spec: v1beta2.BGPPeerSpec{
							Address: "1.2.3.4",
							MyASN:   123,
						},
					},
					{
						Spec: v1beta2.BGPPeerSpec{
							Address: "1.2.3.5",
							MyASN:   123,
						},
					},
					{
						Spec: v1beta2.BGPPeerSpec{
							Address: "1.2.3.6",
							MyASN:   124,
						},
					},
				},
			},
			mustFail: true,
		},
		{
			desc: "myAsn set, one different but with different vrf",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					{
						Spec: v1beta2.BGPPeerSpec{
							Address: "1.2.3.4",
							MyASN:   123,
						},
					},
					{
						Spec: v1beta2.BGPPeerSpec{
							Address: "1.2.3.5",
							MyASN:   123,
						},
					},
					{
						Spec: v1beta2.BGPPeerSpec{
							Address: "1.2.3.6",
							MyASN:   124,
							VRFName: "red",
						},
					},
				},
			},
		},
		{
			desc: "myAsn set, two different but with different vrf",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					{
						Spec: v1beta2.BGPPeerSpec{
							Address: "1.2.3.4",
							MyASN:   123,
							VRFName: "red",
						},
					},
					{
						Spec: v1beta2.BGPPeerSpec{
							Address: "1.2.3.5",
							MyASN:   123,
							VRFName: "red",
						},
					},
					{
						Spec: v1beta2.BGPPeerSpec{
							Address: "1.2.3.6",
							MyASN:   124,
						},
					},
				},
			},
		},
		{
			desc: "duplicate bgp address",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					{
						Spec: v1beta2.BGPPeerSpec{
							Address: "1.2.3.4",
						},
					},
					{
						Spec: v1beta2.BGPPeerSpec{
							Address: "1.2.3.4",
						},
					},
				},
			},
			mustFail: true,
		}, {
			desc: "duplicate bgp address, different vrfs",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					{
						Spec: v1beta2.BGPPeerSpec{
							Address: "1.2.3.4",
						},
					},
					{
						Spec: v1beta2.BGPPeerSpec{
							Address: "1.2.3.4",
							VRFName: "red",
						},
					}, {
						Spec: v1beta2.BGPPeerSpec{
							Address: "1.2.3.4",
							VRFName: "green",
						},
					},
				},
			},
			mustFail: false,
		}, {
			desc: "duplicate bgp address, same vrf",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					{
						Spec: v1beta2.BGPPeerSpec{
							Address: "1.2.3.4",
						},
					},
					{
						Spec: v1beta2.BGPPeerSpec{
							Address: "1.2.3.4",
							VRFName: "red",
						},
					}, {
						Spec: v1beta2.BGPPeerSpec{
							Address: "1.2.3.4",
							VRFName: "red",
						},
					},
				},
			},
			mustFail: true,
		},
		{
			desc: "two peers with interface set different",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					{
						Spec: v1beta2.BGPPeerSpec{
							Interface: "eth0",
							MyASN:     123,
						},
					},
					{
						Spec: v1beta2.BGPPeerSpec{
							Interface: "eth1",
							MyASN:     123,
						},
					},
				},
			},
			mustFail: false,
		},
		{
			desc: "two peers with interface set same",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					{
						Spec: v1beta2.BGPPeerSpec{
							Interface: "eth0",
							MyASN:     123,
						},
					},
					{
						Spec: v1beta2.BGPPeerSpec{
							Interface: "eth0",
							MyASN:     123,
						},
					},
				},
			},
			mustFail: true,
		},
		// Node-selector-aware duplicate detection: same address is only a real
		// duplicate when both peers can be scheduled on the same node.
		{
			desc: "duplicate address, disjoint hostname selectors",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{peerOnNodes("node1"), peerOnNodes("node2")},
				Nodes: []corev1.Node{hostNode("node1"), hostNode("node2")},
			},
		},
		{
			desc: "duplicate address, overlapping hostname selectors",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{peerOnNodes("node1"), peerOnNodes("node1")},
				Nodes: []corev1.Node{hostNode("node1")},
			},
			mustFail: true,
		},
		{
			desc: "duplicate address, one peer without nodeSelector matches all nodes",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					{Spec: v1beta2.BGPPeerSpec{Address: "1.2.3.4"}}, // no selector = all nodes
					peerOnNodes("node1"),
				},
				Nodes: []corev1.Node{hostNode("node1")},
			},
			mustFail: true,
		},
		{
			desc: "duplicate address, disjoint selectors but no nodes known (conservative)",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{peerOnNodes("node1"), peerOnNodes("node2")},
				// No Nodes: conservative fallback must reject.
			},
			mustFail: true,
		},
		{
			desc: "three peers same address, all on distinct nodes",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					peerOnNodes("node1"),
					peerOnNodes("node2"),
					peerOnNodes("node3"),
				},
				Nodes: []corev1.Node{hostNode("node1"), hostNode("node2"), hostNode("node3")},
			},
		},
		{
			desc: "three peers same address, third overlaps with first",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					peerOnNodes("node1"),
					peerOnNodes("node2"),
					peerOnNodes("node1"), // collides with first peer
				},
				Nodes: []corev1.Node{hostNode("node1"), hostNode("node2")},
			},
			mustFail: true,
		},
		{
			desc: "duplicate address, multiple nodeSelectors (OR), disjoint",
			// peer1: node1 OR node2; peer2: node3 OR node4 → no overlap
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					peerOnNodes("node1", "node2"),
					peerOnNodes("node3", "node4"),
				},
				Nodes: []corev1.Node{hostNode("node1"), hostNode("node2"), hostNode("node3"), hostNode("node4")},
			},
		},
		{
			desc: "duplicate address, multiple nodeSelectors (OR), overlap via second selector",
			// peer1: node1 OR node2; peer2: node3 OR node2 → node2 overlaps
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					peerOnNodes("node1", "node2"),
					peerOnNodes("node3", "node2"),
				},
				Nodes: []corev1.Node{hostNode("node1"), hostNode("node2"), hostNode("node3")},
			},
			mustFail: true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			err := DiscardNativeOnly(test.config)
			if test.mustFail && err == nil {
				t.Fatalf("Expected error for %s", test.desc)
			}
			if !test.mustFail && err != nil {
				t.Fatalf("Not expected error %s for %s", err, test.desc)
			}
		})
	}
}

// peerOnNodes returns a BGPPeer for address "1.2.3.4" selecting nodes by hostname (OR semantics).
func peerOnNodes(hostnames ...string) v1beta2.BGPPeer {
	sels := make([]v1.LabelSelector, len(hostnames))
	for i, h := range hostnames {
		sels[i] = v1.LabelSelector{MatchLabels: map[string]string{"kubernetes.io/hostname": h}}
	}
	return v1beta2.BGPPeer{Spec: v1beta2.BGPPeerSpec{Address: "1.2.3.4", NodeSelectors: sels}}
}

// hostNode returns a Node with the given name and kubernetes.io/hostname label.
func hostNode(hostname string) corev1.Node {
	return corev1.Node{ObjectMeta: v1.ObjectMeta{
		Name:   hostname,
		Labels: map[string]string{"kubernetes.io/hostname": hostname},
	}}
}

// TestFRRMyASNNodeSelectorOverlap covers the one-ASN-per-VRF constraint at FRR
// process scope. Peers that cannot be scheduled on the same node never share an
// FRR instance, so different myASN values in one VRF are valid. Peers that can
// land on the same node are still rejected.
func TestFRRMyASNNodeSelectorOverlap(t *testing.T) {
	blue := labelSelector("pool", "blue")
	green := labelSelector("pool", "green")
	shared := labelSelector("pool", "shared")
	colorBlue := labelSelector("color", "blue")
	zoneEast := labelSelector("zone", "east")

	inBlue := v1.LabelSelector{MatchExpressions: []v1.LabelSelectorRequirement{{
		Key:      "pool",
		Operator: v1.LabelSelectorOpIn,
		Values:   []string{"blue"},
	}}}
	inGreen := v1.LabelSelector{MatchExpressions: []v1.LabelSelectorRequirement{{
		Key:      "pool",
		Operator: v1.LabelSelectorOpIn,
		Values:   []string{"green"},
	}}}

	nodes := []corev1.Node{
		labeledNode("node-blue", "pool", "blue"),
		labeledNode("node-green", "pool", "green"),
		labeledNode("node-shared", "pool", "shared"),
	}

	tests := []struct {
		desc     string
		config   ClusterResources
		mustFail bool
	}{
		{
			desc: "different myASN, default VRF, disjoint pool selectors",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					peerWithMyASN("1.2.3.4", 65001, "", blue),
					peerWithMyASN("1.2.3.5", 65002, "", green),
				},
				Nodes: nodes,
			},
		},
		{
			desc: "different myASN, default VRF, overlapping pool selectors",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					peerWithMyASN("1.2.3.4", 65001, "", blue),
					peerWithMyASN("1.2.3.5", 65002, "", blue),
				},
				Nodes: nodes,
			},
			mustFail: true,
		},
		{
			desc: "different myASN, default VRF, one peer selects every node",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					peerWithMyASN("1.2.3.4", 65001, ""),
					peerWithMyASN("1.2.3.5", 65002, "", green),
				},
				Nodes: nodes,
			},
			mustFail: true,
		},
		{
			desc: "different myASN, disjoint selectors, no nodes known",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					peerWithMyASN("1.2.3.4", 65001, "", blue),
					peerWithMyASN("1.2.3.5", 65002, "", green),
				},
			},
			mustFail: true,
		},
		{
			desc: "different myASN, same named VRF, disjoint selectors",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					peerWithMyASN("1.2.3.4", 65001, "red", blue),
					peerWithMyASN("1.2.3.5", 65002, "red", green),
				},
				Nodes: nodes,
			},
		},
		{
			desc: "different myASN, same named VRF, overlapping selectors",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					peerWithMyASN("1.2.3.4", 65001, "red", blue),
					peerWithMyASN("1.2.3.5", 65002, "red", blue),
				},
				Nodes: nodes,
			},
			mustFail: true,
		},
		{
			desc: "different myASN, different VRFs, overlapping selectors",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					peerWithMyASN("1.2.3.4", 65001, "red", blue),
					peerWithMyASN("1.2.3.5", 65002, "green", blue),
				},
				Nodes: nodes,
			},
		},
		{
			desc: "same myASN, default VRF, overlapping selectors",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					peerWithMyASN("1.2.3.4", 65001, "", blue),
					peerWithMyASN("1.2.3.5", 65001, "", blue),
				},
				Nodes: nodes,
			},
		},
		{
			desc: "different myASN, same address, disjoint selectors",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					peerWithMyASN("1.2.3.4", 65001, "", blue),
					peerWithMyASN("1.2.3.4", 65002, "", green),
				},
				Nodes: nodes,
			},
		},
		{
			desc: "different myASN only between peers that do not share a node",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					peerWithMyASN("1.2.3.4", 65001, "", blue),
					peerWithMyASN("1.2.3.5", 65002, "", green),
					peerWithMyASN("1.2.3.6", 65001, "", blue),
				},
				Nodes: nodes,
			},
		},
		{
			desc: "different myASN on a shared node, with a third disjoint peer",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					peerWithMyASN("1.2.3.4", 65001, "", blue),
					peerWithMyASN("1.2.3.5", 65002, "", green),
					peerWithMyASN("1.2.3.6", 65003, "", blue),
				},
				Nodes: nodes,
			},
			mustFail: true,
		},
		{
			desc: "different myASN, OR selectors overlap on the second selector",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					peerWithMyASN("1.2.3.4", 65001, "", blue, shared),
					peerWithMyASN("1.2.3.5", 65002, "", green, shared),
				},
				Nodes: nodes,
			},
			mustFail: true,
		},
		{
			desc: "different myASN, matchExpressions select disjoint pools",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					peerWithMyASN("1.2.3.4", 65001, "", inBlue),
					peerWithMyASN("1.2.3.5", 65002, "", inGreen),
				},
				Nodes: nodes,
			},
		},
		{
			desc: "different myASN, selectors on different keys match one node",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					peerWithMyASN("1.2.3.4", 65001, "", colorBlue),
					peerWithMyASN("1.2.3.5", 65002, "", zoneEast),
				},
				Nodes: []corev1.Node{{
					ObjectMeta: v1.ObjectMeta{
						Name: "node-both",
						Labels: map[string]string{
							"color": "blue",
							"zone":  "east",
						},
					},
				}},
			},
			mustFail: true,
		},
		{
			desc: "different myASN, selectors on different keys match no common node",
			config: ClusterResources{
				Peers: []v1beta2.BGPPeer{
					peerWithMyASN("1.2.3.4", 65001, "", colorBlue),
					peerWithMyASN("1.2.3.5", 65002, "", zoneEast),
				},
				Nodes: []corev1.Node{
					{ObjectMeta: v1.ObjectMeta{Name: "node-a", Labels: map[string]string{"color": "blue", "zone": "west"}}},
					{ObjectMeta: v1.ObjectMeta{Name: "node-b", Labels: map[string]string{"color": "red", "zone": "east"}}},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			err := DiscardNativeOnly(test.config)
			if test.mustFail {
				if err == nil {
					t.Fatalf("expected myASN error for %s", test.desc)
				}
				if !strings.Contains(err.Error(), "myAsn") {
					t.Fatalf("expected myASN error for %s, got %v", test.desc, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %s: %v", test.desc, err)
			}
		})
	}
}

// TestValidatorHonorsMyASNNodeOverlap checks the admission path: NodeList has to
// be part of the validated objects, otherwise disjoint selectors cannot be
// proven and the myASN check stays conservative.
func TestValidatorHonorsMyASNNodeOverlap(t *testing.T) {
	blue := labelSelector("pool", "blue")
	green := labelSelector("pool", "green")
	peers := &v1beta2.BGPPeerList{Items: []v1beta2.BGPPeer{
		namedPeerWithMyASN("peer-blue", "192.0.2.1", 65001, 64500, "", blue),
		namedPeerWithMyASN("peer-green", "192.0.2.2", 65002, 64500, "", green),
	}}
	nodes := &corev1.NodeList{Items: []corev1.Node{
		labeledNode("node-blue", "pool", "blue"),
		labeledNode("node-green", "pool", "green"),
	}}

	v := NewValidator(DiscardNativeOnly)
	if err := v.Validate(peers); err == nil || !strings.Contains(err.Error(), "myAsn") {
		t.Fatalf("expected myASN rejection without a node list, got %v", err)
	}
	if err := v.Validate(peers, nodes); err != nil {
		t.Fatalf("disjoint peers should be accepted when the node list is validated: %v", err)
	}

	overlap := peers.DeepCopy()
	overlap.Items[1].Spec.NodeSelectors = append([]v1.LabelSelector(nil), peers.Items[0].Spec.NodeSelectors...)
	if err := v.Validate(overlap, nodes); err == nil || !strings.Contains(err.Error(), "myAsn") {
		t.Fatalf("expected myASN rejection for overlapping selectors, got %v", err)
	}
}

func peerWithMyASN(address string, myASN uint32, vrf string, selectors ...v1.LabelSelector) v1beta2.BGPPeer {
	return v1beta2.BGPPeer{Spec: v1beta2.BGPPeerSpec{
		Address:       address,
		MyASN:         myASN,
		VRFName:       vrf,
		NodeSelectors: selectors,
	}}
}

func namedPeerWithMyASN(name, address string, myASN, asn uint32, vrf string, selectors ...v1.LabelSelector) v1beta2.BGPPeer {
	peer := peerWithMyASN(address, myASN, vrf, selectors...)
	peer.Name = name
	peer.Spec.ASN = asn
	return peer
}

func labelSelector(key, value string) v1.LabelSelector {
	return v1.LabelSelector{MatchLabels: map[string]string{key: value}}
}

func labeledNode(name, key, value string) corev1.Node {
	return corev1.Node{ObjectMeta: v1.ObjectMeta{
		Name:   name,
		Labels: map[string]string{key: value},
	}}
}
