package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/elb"
)

func (m *handler) getELBV1s(clusterName string, vpcID string) ([]string, error) {
	elbsInVPC, err := m.getELBV1NamesInVPC(vpcID)
	if err != nil {
		return nil, err
	}

	expectedTag := fmt.Sprintf("kubernetes/cluster/%s", clusterName)
	filteredELBs, err := m.filterELBV1sWithTag(elbsInVPC, expectedTag)
	if err != nil {
		return nil, err
	}
	return filteredELBs, nil
}

func (m *handler) filterELBV1sWithTag(elbNames []*string, tagName string) ([]string, error) {
	elbTags, err := m.ELB.DescribeTags(&elb.DescribeTagsInput{
		LoadBalancerNames: elbNames})
	if err != nil {
		return nil, err
	}

	names := []string{}
	for _, element := range elbTags.TagDescriptions {
		for _, tag := range element.Tags {
			if *tag.Key == tagName {
				names = append(names, *element.LoadBalancerName)
				break
			}
		}
	}

	return names, nil
}

func (m *handler) getELBV1NamesInVPC(vpcID string) ([]*string, error) {
	elbDescribeParams := &elb.DescribeLoadBalancersInput{}
	elbs, err := m.ELB.DescribeLoadBalancers(elbDescribeParams)
	if err != nil {
		return nil, err
	}

	elbsInVPC := []*string{}
	for _, element := range elbs.LoadBalancerDescriptions {
		if *element.VPCId == vpcID {
			elbsInVPC = append(elbsInVPC, element.LoadBalancerName)
		}
	}
	return elbsInVPC, nil
}

func (m *handler) drainNodeFromELBV1(nodeID string, elbV1Name string) (done bool, e error) {
	result, err := m.ELB.DescribeInstanceHealth(&elb.DescribeInstanceHealthInput{
		LoadBalancerName: &elbV1Name})
	if err != nil {
		return false, err
	}

	instanceAtELB := false
	for _, element := range result.InstanceStates {
		if *element.InstanceId == nodeID && *element.State == "InService" {
			instanceAtELB = true
			break
		}
	}

	if !instanceAtELB {
		return true, nil
	}

	_, err = m.ELB.DeregisterInstancesFromLoadBalancer(&elb.DeregisterInstancesFromLoadBalancerInput{
		Instances:        []*elb.Instance{&elb.Instance{InstanceId: &nodeID}},
		LoadBalancerName: &elbV1Name})
	if err != nil {
		return false, err
	}

	return false, nil
}