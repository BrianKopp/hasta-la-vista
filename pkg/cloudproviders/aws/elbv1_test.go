package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
)

func TestDescribeLoadBalancers(t *testing.T) {
	clients := CloudProvider{
		ELB: &fakeELB{
			describeELBOutput: &elb.DescribeLoadBalancersOutput{
				LoadBalancerDescriptions: []*elb.LoadBalancerDescription{
					&elb.LoadBalancerDescription{
						LoadBalancerName: aws.String("ELBA"),
						VPCId:            aws.String("vpc-1"),
					},
					&elb.LoadBalancerDescription{
						LoadBalancerName: aws.String("ELBB"),
						VPCId:            aws.String("vpc-2"),
					},
				},
			},
		},
	}
	elbs, _ := clients.getELBV1NamesInVPC("vpc-1")
	if len(elbs) != 1 {
		t.Fatalf("Expected only one result, got %v", len(elbs))
	}
	if *elbs[0] != "ELBA" {
		t.Fatalf("Expected name ELBA, got %v", elbs[0])
	}
	return
}

func TestFilterELBV1s(t *testing.T) {
	cases := []struct {
		Resp     elb.DescribeTagsOutput
		Expected []string
	}{
		{
			Resp: elb.DescribeTagsOutput{
				TagDescriptions: []*elb.TagDescription{
					&elb.TagDescription{
						LoadBalancerName: aws.String("ELB1"),
						Tags: []*elb.Tag{
							&elb.Tag{
								Key:   aws.String("WrongKey"),
								Value: aws.String("matching"),
							},
						},
					},
				},
			},
			Expected: []string{},
		},
		{
			Resp: elb.DescribeTagsOutput{
				TagDescriptions: []*elb.TagDescription{
					&elb.TagDescription{
						LoadBalancerName: aws.String("ELB1"),
						Tags: []*elb.Tag{
							&elb.Tag{
								Key:   aws.String("WrongKey"),
								Value: aws.String("matching"),
							},
							&elb.Tag{
								Key:   aws.String("kubernetets.io/cluster/clustername"),
								Value: aws.String("matching"),
							},
						},
					},
					&elb.TagDescription{
						LoadBalancerName: aws.String("ELB2"),
						Tags: []*elb.Tag{
							&elb.Tag{
								Key:   aws.String("WrongKey"),
								Value: aws.String("matching"),
							},
						},
					},
				},
			},
			Expected: []string{"ELB1"},
		},
	}

	for i, c := range cases {
		clients := &CloudProvider{
			ELB: &fakeELB{
				describeTagsOutput: &c.Resp,
			},
		}
		filtered, _ := clients.filterELBV1sWithTag(nil, "kubernetets.io/cluster/clustername")
		if len(filtered) != len(c.Expected) {
			t.Fatalf("%d failed - unexpected number of results, expected %v, actual %v", i, len(c.Expected), len(filtered))
		}
		for j, n := range c.Expected {
			if n != filtered[j] {
				t.Fatalf("%d failed - did not match %d index of expected matching list. expected %v, actual %v", i, j, n, filtered[j])
			}
		}
	}
}

func TestDrainNodeFromELBV1WhenPresent(t *testing.T) {
	clients := &CloudProvider{
		ELB: &fakeELB{
			descHealthOutput: &elb.DescribeInstanceHealthOutput{
				InstanceStates: []*elb.InstanceState{
					&elb.InstanceState{
						InstanceId: aws.String("myinstance"),
						State:      aws.String("InService"),
					},
				},
			},
			deregOutput: &elb.DeregisterInstancesFromLoadBalancerOutput{
				Instances: []*elb.Instance{},
			},
		},
	}
	result, err := clients.drainNodeFromELBV1("myinstance", "")
	if result || err != nil {
		t.Fatalf("failed - expected result to be false and err to be nil")
	}
}

func TestDrainNodeFromELBV1WhenNotPresent(t *testing.T) {
	clients := &CloudProvider{
		ELB: &fakeELB{
			descHealthOutput: &elb.DescribeInstanceHealthOutput{
				InstanceStates: []*elb.InstanceState{
					&elb.InstanceState{
						InstanceId: aws.String("myinstance"),
						State:      aws.String("InService"),
					},
				},
			},
			deregOutput: &elb.DeregisterInstancesFromLoadBalancerOutput{
				Instances: []*elb.Instance{},
			},
		},
	}
	result, err := clients.drainNodeFromELBV1("differentinstance", "")
	if err != nil {
		t.Fatalf("failed - expected err to be nil")
	}
	if result != true {
		t.Fatalf("failed - expected result to be true")
	}
}
