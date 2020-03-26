package main

import (
	"github.com/aws/aws-sdk-go/service/elb"
)

type fakeELB struct {
	describeELBOutput  *elb.DescribeLoadBalancersOutput
	describeTagsOutput *elb.DescribeTagsOutput
	descHealthOutput   *elb.DescribeInstanceHealthOutput
	deregOutput        *elb.DeregisterInstancesFromLoadBalancerOutput
	err                error
}

func (m *fakeELB) DescribeLoadBalancers(input *elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error) {
	return m.describeELBOutput, nil
}

func (m *fakeELB) DescribeTags(input *elb.DescribeTagsInput) (*elb.DescribeTagsOutput, error) {
	return m.describeTagsOutput, nil
}

func (m *fakeELB) DeregisterInstancesFromLoadBalancer(input *elb.DeregisterInstancesFromLoadBalancerInput) (*elb.DeregisterInstancesFromLoadBalancerOutput, error) {
	return m.deregOutput, nil
}
func (m *fakeELB) DescribeInstanceHealth(input *elb.DescribeInstanceHealthInput) (*elb.DescribeInstanceHealthOutput, error) {
	return m.descHealthOutput, nil
}
