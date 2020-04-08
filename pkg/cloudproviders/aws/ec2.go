package aws

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// GetVPCAndClusterFromInstance gets the VPC ID and cluster name from the instance
func (m *CloudProvider) GetVPCAndClusterFromInstance(nodeID string) (vpcID *string, clusterName *string, err error) {
	instances, err := m.EC2.DescribeInstances(
		&ec2.DescribeInstancesInput{
			InstanceIds: []*string{
				aws.String(nodeID),
			},
		},
	)

	if err != nil {
		return nil, nil, err
	}

	tagKeyMatch := "kubernetes.io/cluster/"
	for _, res := range instances.Reservations {
		for _, inst := range res.Instances {
			vpcID := inst.VpcId
			for _, tagPair := range inst.Tags {
				if strings.HasPrefix(*tagPair.Key, tagKeyMatch) {
					clusterName := (*tagPair.Key)[(len(tagKeyMatch)):]
					return vpcID, &clusterName, nil
				}
			}
			return nil, nil, errors.New("Could not find matching tag on instance")
		}
	}

	return nil, nil, nil
}

// getNodeIDFromIP gets the node id from the ip...
func (m *CloudProvider) getNodeIDFromIP(nodeIP string) (*string, error) {
	instances, err := m.EC2.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("private-ip-address"),
				Values: []*string{&nodeIP},
			},
		},
	})

	if err != nil {
		return nil, err
	}

	for _, res := range instances.Reservations {
		for _, inst := range res.Instances {
			return inst.InstanceId, nil
		}
	}

	return nil, errors.New("Unable to find instance by IP")
}
