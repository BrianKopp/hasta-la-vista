package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elb"
)

// clientFactory is an interface used to acquire aws clients
type clientFactory struct {
	Session *session.Session
	region  string
}

// getELBClient gets an elb client
func (f *clientFactory) getELBClient() *elb.ELB {
	config := aws.Config{Region: aws.String(f.region)}
	return elb.New(f.Session, &config)
}

type myELBAPI interface {
	DeregisterInstancesFromLoadBalancer(input *elb.DeregisterInstancesFromLoadBalancerInput) (*elb.DeregisterInstancesFromLoadBalancerOutput, error)
	DescribeInstanceHealth(input *elb.DescribeInstanceHealthInput) (*elb.DescribeInstanceHealthOutput, error)
	DescribeLoadBalancers(input *elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error)
	DescribeTags(input *elb.DescribeTagsInput) (*elb.DescribeTagsOutput, error)
}

type handler struct {
	ELB myELBAPI
}
