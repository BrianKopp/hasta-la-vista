package main

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"

	"github.com/rs/zerolog/log"
)

var awsSession = session.Must(session.NewSession())
var elbClient *elb.ELB
var elbV2Client *elbv2.ELBV2

func handleDeregistration(nIP string, nID string, clusterName string, vpcID string) error {
	log.Info().
		Str("nodeIP", nIP).
		Str("nodeID", nID).
		Str("clusterName", clusterName).
		Str("vpcID", vpcID).
		Msg("Handling deregistration for node")
	if nID == "" && nIP == "" {
		return errors.New("nodeIP and nodeID cannot both be empty")
	}

	if clusterName == "" {
		return errors.New("clusterName cannot be empty")
	}

	if vpcID == "" {
		return errors.New("vpcID cannot be empty")
	}

	var nodeID string
	if nID == "" {
		id, err := getNodeIDFromIP(nIP)
		if err != nil {
			return err
		}

		log.Info().Str("nodeID", id).Msg("acquired node id from node ip")
		nodeID = id
	}

	var wg sync.WaitGroup
	log.Info().Msg("beginning drain operations")
	go func() {
		err := drainNodeFromELBV1sInCluster(nodeID, clusterName, vpcID)
		if err != nil {
			log.Error().
				Err(err).
				Str("nodeID", nodeID).
				Msg("error occurred draining node from all v1 ELBs")
			wg.Done()
			return
		}
		log.Info().Str("nodeID", nodeID).Msg("successfully drained from all v1 ELBs")
		wg.Done()
	}()

	go func() {
		err := drainNodeFromELBV2sInCluster(nodeID, clusterName, vpcID)
		if err != nil {
			log.Error().
				Err(err).
				Str("nodeID", nodeID).
				Msg("error occurred draining node from all v2 ELBs")
			wg.Done()
			return
		}
		log.Info().Str("nodeID", nodeID).Msg("successfully drained from all v2 ELBs")
		wg.Done()
	}()

	wg.Wait()
	return nil // TODO report error
}

// getNodeIDFromIP gets the node id from the ip...
func getNodeIDFromIP(nodeIP string) (string, error) {
	return "TODO", nil
}

func drainNodeFromELBV1sInCluster(nodeID string, clusterName string, vpcID string) error {
	elbV1Names, err := getELBV1NamesInCluster(clusterName, vpcID)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, elbV1Name := range elbV1Names {
		wg.Add(1)
		start := time.Now()
		go func(name string) {
			for {
				drained, _ := drainNodeFromELBV1(nodeID, name)
				if drained {
					wg.Done()
					break
				}
				if time.Since(start) > (120 * time.Second) {
					wg.Done()
					break
				}
				time.Sleep(5 * time.Second)
			}
		}(elbV1Name)
	}
	wg.Wait()
	return nil
}

func drainNodeFromELBV2sInCluster(nodeID string, clusterName string, vpcID string) error {
	targetGroupARNs, err := getELBV2TargetGroupARNsInCluster(clusterName, vpcID)
	if err != nil {
		return nil
	}

	var wg sync.WaitGroup
	for _, targetGroupARN := range targetGroupARNs {
		wg.Add(1)
		start := time.Now()
		go func(arn string) {
			for {
				drained, _ := drainNodeFromELBV2TargetGroup(nodeID, arn)
				if drained {
					wg.Done()
					break
				}
				if time.Since(start) > (120 * time.Second) {
					wg.Done()
					break
				}
				time.Sleep(5 * time.Second)
			}
		}(targetGroupARN)
	}

	wg.Wait()
	return nil
}

func getELBV1NamesInCluster(clusterName string, vpcID string) ([]string, error) {
	elbDescribeParams := &elb.DescribeLoadBalancersInput{}
	elbClient := getELBClient()
	elbs, err := elbClient.DescribeLoadBalancers(elbDescribeParams)
	if err != nil {
		return nil, err
	}

	elbsInVPC := []*string{}
	for _, element := range elbs.LoadBalancerDescriptions {
		if *element.VPCId == vpcID {
			elbsInVPC = append(elbsInVPC, element.LoadBalancerName)
		}
	}

	elbTags, err := elbClient.DescribeTags(&elb.DescribeTagsInput{
		LoadBalancerNames: elbsInVPC})
	if err != nil {
		return nil, err
	}

	expectedTag := fmt.Sprintf("kubernetes/cluster/%s", clusterName)
	names := []string{}
	for _, element := range elbTags.TagDescriptions {
		for _, tag := range element.Tags {
			if *tag.Key == expectedTag {
				names = append(names, *element.LoadBalancerName)
				break
			}
		}
	}

	return names, nil
}

func (m *awsClients) getELBV1NamesInCluster(clusterName string, vpcID string) ([]string, error) {
	elbsInVPC, err := m.getELBV1NamesInVPC(vpcID)
	if err != nil {
		return nil, err
	}

	elbTags, err := m.ELB.DescribeTags(&elb.DescribeTagsInput{
		LoadBalancerNames: elbsInVPC})
	if err != nil {
		return nil, err
	}

	expectedTag := fmt.Sprintf("kubernetes/cluster/%s", clusterName)
	names := []string{}
	for _, element := range elbTags.TagDescriptions {
		for _, tag := range element.Tags {
			if *tag.Key == expectedTag {
				names = append(names, *element.LoadBalancerName)
				break
			}
		}
	}

	return names, nil
}

func (m *awsClients) getELBV1NamesInVPC(vpcID string) ([]*string, error) {
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

func getELBV2TargetGroupARNsInCluster(clusterName string, vpcID string) ([]string, error) {
	elbV2DescribeParams := &elbv2.DescribeLoadBalancersInput{}
	elbV2Client := getELBV2Client()

	elbs, err := elbV2Client.DescribeLoadBalancers(elbV2DescribeParams)
	if err != nil {
		return nil, err
	}

	elbsInVPC := []*string{}
	for _, element := range elbs.LoadBalancers {
		if *element.VpcId == vpcID {
			elbsInVPC = append(elbsInVPC, element.LoadBalancerArn)
		}
	}

	elbTags, err := elbV2Client.DescribeTags(&elbv2.DescribeTagsInput{
		ResourceArns: elbsInVPC})
	if err != nil {
		return nil, err
	}

	expectedTag := fmt.Sprintf("kubernetes/cluster/%s", clusterName)
	elbV2ARNs := []*string{}
	for _, element := range elbTags.TagDescriptions {
		for _, tag := range element.Tags {
			if *tag.Key == expectedTag {
				elbV2ARNs = append(elbV2ARNs, element.ResourceArn)
				break
			}
		}
	}

	targetGroupARNs := []string{}
	for _, elbARN := range elbV2ARNs {
		listeners, err := elbV2Client.DescribeListeners(&elbv2.DescribeListenersInput{
			LoadBalancerArn: elbARN})
		if err != nil {
			return nil, err
		}
		for _, listener := range listeners.Listeners {
			if len(listener.DefaultActions) > 0 {
				da := listener.DefaultActions[0]
				if !contains(targetGroupARNs, *da.TargetGroupArn) {
					targetGroupARNs = append(targetGroupARNs, *da.TargetGroupArn)
				}
			}
		}
	}

	return targetGroupARNs, nil
}

func drainNodeFromELBV1(nodeID string, elbV1Name string) (done bool, e error) {
	elbClient := getELBClient()
	result, err := elbClient.DescribeInstanceHealth(&elb.DescribeInstanceHealthInput{
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

	_, err = elbClient.DeregisterInstancesFromLoadBalancer(&elb.DeregisterInstancesFromLoadBalancerInput{
		Instances:        []*elb.Instance{&elb.Instance{InstanceId: &nodeID}},
		LoadBalancerName: &elbV1Name})
	if err != nil {
		return false, err
	}

	return false, nil
}

func drainNodeFromELBV2TargetGroup(nodeID string, targetGroupArn string) (done bool, e error) {
	elbV2Client := getELBV2Client()
	healthResult, err := elbV2Client.DescribeTargetHealth(&elbv2.DescribeTargetHealthInput{
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

	if !instanceAtELB {
		return true, nil
	}

	requireDraining := state != "draining"
	if requireDraining {
		_, err = elbV2Client.DeregisterTargets(&elbv2.DeregisterTargetsInput{
			TargetGroupArn: &targetGroupArn,
			Targets:        []*elbv2.TargetDescription{&elbv2.TargetDescription{Id: &nodeID}}})
		if err != nil {
			return false, err
		}
	}

	return false, nil
}

func getELBClient() *elb.ELB {
	if elbClient == nil {
		config := aws.Config{
			Region: aws.String(getAWSRegion()),
		}
		elbClient = elb.New(awsSession, &config)
	}

	return elbClient
}

func getELBV2Client() *elbv2.ELBV2 {
	if elbV2Client == nil {
		config := aws.Config{
			Region: aws.String(getAWSRegion()),
		}
		elbV2Client = elbv2.New(awsSession, &config)
	}

	return elbV2Client
}

func contains(lst []string, s string) bool {
	for _, a := range lst {
		if a == s {
			return true
		}
	}
	return false
}
