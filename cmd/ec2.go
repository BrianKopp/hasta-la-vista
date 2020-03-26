package main

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// getNodeIDFromIP gets the node id from the ip...
func (m *handler) getNodeIDFromIP(nodeIP string) (*string, error) {
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
