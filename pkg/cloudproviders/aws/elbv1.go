package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/rs/zerolog/log"
)

func (m *CloudProvider) getELBV1s(vpcID string, clusterName string) ([]string, error) {
	elbsInVPC, err := m.getELBV1NamesInVPC(vpcID)
	if err != nil {
		return nil, err
	}

	expectedTag := fmt.Sprintf("kubernetes.io/cluster/%s", clusterName)
	filteredELBs, err := m.filterELBV1sWithTag(elbsInVPC, expectedTag)
	if err != nil {
		return nil, err
	}

	log.Debug().
		Str("elbNames", fmt.Sprintf("%v", filteredELBs)).
		Str("vpcID", vpcID).
		Str("clusterName", clusterName).
		Msg("retrieved elbv1s in vpc for cluster name")
	return filteredELBs, nil
}

func (m *CloudProvider) filterELBV1sWithTag(elbNames []*string, tagName string) ([]string, error) {
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

	log.Debug().
		Int("preFilterList", len(elbNames)).
		Int("postFilterList", len(names)).
		Str("filteredNames", fmt.Sprintf("%v", names)).
		Str("tagName", tagName).
		Msg("filtered elb list by tagname")
	return names, nil
}

func (m *CloudProvider) getELBV1NamesInVPC(vpcID string) ([]*string, error) {
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

	log.Debug().
		Str("vpcID", vpcID).
		Str("elbNames", fmt.Sprintf("%v", elbsInVPC)).
		Msg("found elb v1s in vpc")
	return elbsInVPC, nil
}

func (m *CloudProvider) drainNodeFromELBV1(nodeID string, elbV1Name string) (done bool, e error) {
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
		log.Info().
			Str("nodeID", nodeID).
			Str("elbName", elbV1Name).
			Msg("Instance not InService at ELB")
		return true, nil
	}

	if m.DryRun {
		log.Info().
			Str("nodeID", nodeID).
			Str("elbName", elbV1Name).
			Msg("DRY-RUN (no action taken)---Node InService at elb, draining (faking success)")
		return false, nil
	}

	log.Info().
		Str("nodeID", nodeID).
		Str("elbName", elbV1Name).
		Msg("Node InService at elb, draining")

	_, err = m.ELB.DeregisterInstancesFromLoadBalancer(&elb.DeregisterInstancesFromLoadBalancerInput{
		Instances:        []*elb.Instance{&elb.Instance{InstanceId: &nodeID}},
		LoadBalancerName: &elbV1Name,
	})

	if err != nil {
		return false, err
	}

	return false, nil
}
