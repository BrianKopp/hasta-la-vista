package main

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
)

type fakeELB struct {
	describeELBOutput  *elb.DescribeLoadBalancersOutput
	describeTagsOutput *elb.DescribeTagsOutput
}

func (m *fakeELB) DescribeLoadBalancers(input *elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error) {
	return m.describeELBOutput, nil
}

func (m *fakeELB) DescribeTags(input *elb.DescribeTagsInput) (*elb.DescribeTagsOutput, error) {
	return m.describeTagsOutput, nil
}

func TestDescribeLoadBalancers(t *testing.T) {
	clients := awsClients{ELB: &fakeELB{
		describeELBOutput: &elb.DescribeLoadBalancersOutput{
			LoadBalancerDescriptions: []*elb.LoadBalancerDescription{
				&elb.LoadBalancerDescription{
					LoadBalancerName: aws.String("ELBA"),
					VPCId:            aws.String("vpc-1")},
				&elb.LoadBalancerDescription{
					LoadBalancerName: aws.String("ELBB"),
					VPCId:            aws.String("vpc-2")},
			}}}}
	elbs, _ := clients.getELBV1NamesInVPC("vpc-1")
	if len(elbs) != 1 {
		t.Fatalf("Expected only one result, got %v", len(elbs))
	}
	if *elbs[0] != "ELBA" {
		t.Fatalf("Expected name ELBA, got %v", elbs[0])
	}
	return
}
