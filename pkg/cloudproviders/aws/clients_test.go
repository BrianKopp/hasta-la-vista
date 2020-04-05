package aws

import (
	"github.com/aws/aws-sdk-go/service/ec2"
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
	return m.describeELBOutput, m.err
}

func (m *fakeELB) DescribeTags(input *elb.DescribeTagsInput) (*elb.DescribeTagsOutput, error) {
	return m.describeTagsOutput, m.err
}

func (m *fakeELB) DeregisterInstancesFromLoadBalancer(input *elb.DeregisterInstancesFromLoadBalancerInput) (*elb.DeregisterInstancesFromLoadBalancerOutput, error) {
	return m.deregOutput, m.err
}
func (m *fakeELB) DescribeInstanceHealth(input *elb.DescribeInstanceHealthInput) (*elb.DescribeInstanceHealthOutput, error) {
	return m.descHealthOutput, m.err
}

type fakeEC2 struct {
	describeInstancesOutput *ec2.DescribeInstancesOutput
	err                     error
}

func (m *fakeEC2) DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error) {
	return m.describeInstancesOutput, m.err
}
