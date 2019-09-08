package config

import (
	"gopkg.in/yaml.v2"
	"testing"
)

func TestGetNodeType(t *testing.T) {
	type args struct {
		nodeType string
	}
	tests := []struct {
		name string
		args args
		want NodeType
	}{

		{
			name: "correct type-1",
			args: args{"Controller"},
			want: ControlType,
		},
		{
			name: "correct type-2",
			args: args{"Worker"},
			want: WorkerType,
		},
		{
			name: "correct type-3",
			args: args{"Ingress"},
			want: IngressType,
		},
		{
			name: "incorect",
			args: args{"bogus"},
			want: Unknown,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetNodeType(tt.args.nodeType); got != tt.want {
				t.Errorf("GetNodeType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createTestNode(t *testing.T) NodeTemplate {

	var data = `
      prefix: controller
      domainSuffix: vmwarelab.edu
      desiredCount: 3
      desiredAddress: 172.16.81.128/24
      gateway: 172.16.81.100
      vmTemplateName: ubuntu19-template
      clusterName: mgmt
      logicalSwitch: "test-segment"
`
	var testNode = NodeTemplate{}
	err := yaml.Unmarshal([]byte(data), &testNode)
	if err != nil {
		t.Fatal("Bad yaml", err)
	}

	testNode.Mac = append(testNode.Mac, "mac1")
	testNode.Mac = append(testNode.Mac, "mac2")

	testNode.NetworksRef = append(testNode.NetworksRef, "ref1")
	testNode.NetworksRef = append(testNode.NetworksRef, "ref2")

	return testNode
}

func TestNodeTemplate_Clone(t *testing.T) {

	type args struct {
		node NodeTemplate
	}

	tests := []struct {
		name string
		args args
		//		want   *NodeTemplate
	}{
		{
			name: "correct type-1",
			args: args{createTestNode(t)},
			//want: ControlType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cloned := tt.args.node.Clone()
			t.Log(cloned)

			// TODO finish

		})
	}
}
