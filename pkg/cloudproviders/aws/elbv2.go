package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/elbv2"
)

func (m *CloudProvider) getELBV2TargetGroupARNsInCluster() ([]string, error) {
	elbsInVPC, err := m.getELBV2sInVPC()
	if err != nil {
		return nil, err
	}

	expectedTag := fmt.Sprintf("kubernetes/cluster/%s", m.ClusterName)
	filteredELBs, err := m.filterELBV2sWithTag(elbsInVPC, expectedTag)
	if err != nil {
		return nil, err
	}

	targetGroupARNs, err := m.getTargetGroupsAtELBARNs(filteredELBs)
	if err != nil {
		return nil, err
	}

	return targetGroupARNs, nil
}

func (m *CloudProvider) getTargetGroupsAtELBARNs(elbV2ARNs []*string) ([]string, error) {
	targetGroupARNs := []string{}
	for _, elbARN := range elbV2ARNs {
		elbTargets, err := m.getTargetGroupsAtELB(elbARN)
		if err != nil {
			return nil, err
		}
		for _, target := range elbTargets {
			if !contains(targetGroupARNs, *target) {
				targetGroupARNs = append(targetGroupARNs, *target)
			}
		}
	}
	return targetGroupARNs, nil
}

func (m *CloudProvider) getTargetGroupsAtELB(elbV2ARN *string) ([]*string, error) {
	listeners, err := m.ELBV2.DescribeListeners(&elbv2.DescribeListenersInput{
		LoadBalancerArn: elbV2ARN})
	if err != nil {
		return nil, err
	}
	targets := []*string{}
	for _, listener := range listeners.Listeners {
		if len(listener.DefaultActions) > 0 {
			da := listener.DefaultActions[0]
			targets = append(targets, da.TargetGroupArn)
		}
	}
	return targets, nil
}

func (m *CloudProvider) getELBV2sInVPC() ([]*string, error) {
	elbs, err := m.ELBV2.DescribeLoadBalancers(&elbv2.DescribeLoadBalancersInput{})
	if err != nil {
		return nil, err
	}

	elbsInVPC := []*string{}
	for _, element := range elbs.LoadBalancers {
		if *element.VpcId == m.VPCID {
			elbsInVPC = append(elbsInVPC, element.LoadBalancerArn)
		}
	}
	return elbsInVPC, nil
}

func (m *CloudProvider) filterELBV2sWithTag(elbV2ARNs []*string, expectedTag string) ([]*string, error) {
	elbTags, err := m.ELBV2.DescribeTags(&elbv2.DescribeTagsInput{
		ResourceArns: elbV2ARNs,
	})
	if err != nil {
		return nil, err
	}
	filteredARNs := []*string{}
	for _, element := range elbTags.TagDescriptions {
		for _, tag := range element.Tags {
			if *tag.Key == expectedTag {
				filteredARNs = append(filteredARNs, element.ResourceArn)
				break
			}
		}
	}
	return filteredARNs, nil
}

func (m *CloudProvider) drainNodeFromELBV2TargetGroup(nodeID string, targetGroupArn string) (bool, error) {
	needsDraining, err := m.instanceNeedsDrainingFromTargetGroup(nodeID, targetGroupArn)
	if err != nil {
		return false, err
	}

	if needsDraining {
		_, err = m.ELBV2.DeregisterTargets(&elbv2.DeregisterTargetsInput{
			TargetGroupArn: &targetGroupArn,
			Targets:        []*elbv2.TargetDescription{&elbv2.TargetDescription{Id: &nodeID}}})
		if err != nil {
			return false, err
		}
	}

	return false, nil
}

func (m *CloudProvider) instanceNeedsDrainingFromTargetGroup(nodeID string, targetGroupArn string) (bool, error) {
	healthResult, err := m.ELBV2.DescribeTargetHealth(&elbv2.DescribeTargetHealthInput{
		TargetGroupArn: &targetGroupArn})
	if err != nil {
		return false, err
	}

	instanceAtELB := false
	matchingStates := []string{"initial", "healthy", "draining"}
	var state string
	for _, desc := range healthResult.TargetHealthDescriptions {
		if *desc.Target.Id == nodeID && contains(matchingStates, *desc.TargetHealth.State) {
			instanceAtELB = true
			state = *desc.TargetHealth.State
			break
		}
	}

	if instanceAtELB && state != "draining" {
		return true, nil
	}
	return false, nil
}

func contains(lst []string, s string) bool {
	for _, a := range lst {
		if a == s {
			return true
		}
	}
	return false
}
