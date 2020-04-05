package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
)

func TestGetNodeIDFromIP(t *testing.T) {
	expected := "i-0123456789"
	clients := CloudProvider{
		EC2: &fakeEC2{
			describeInstancesOutput: &ec2.DescribeInstancesOutput{
				Reservations: []*ec2.Reservation{
					&ec2.Reservation{
						Instances: []*ec2.Instance{
							&ec2.Instance{
								InstanceId: &expected,
							},
						},
					},
				},
			},
		},
	}

	id, _ := clients.getNodeIDFromIP("ip-10-0-0-1.ec2.internal")
	if *id != expected {
		t.Fatalf("Expected instance id to equal %v, got %v", expected, id)
	}
	return
}
