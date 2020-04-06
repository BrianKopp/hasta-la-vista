package aws

import (
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/rs/zerolog/log"
)

// MyEC2API is a subset of the AWS EC2 API interface
type MyEC2API interface {
	DescribeInstances(*ec2.DescribeInstancesInput) (*ec2.DescribeInstancesOutput, error)
}

// MyELBAPI is a subset of the AWS ELB API interface
type MyELBAPI interface {
	DeregisterInstancesFromLoadBalancer(input *elb.DeregisterInstancesFromLoadBalancerInput) (*elb.DeregisterInstancesFromLoadBalancerOutput, error)
	DescribeInstanceHealth(input *elb.DescribeInstanceHealthInput) (*elb.DescribeInstanceHealthOutput, error)
	DescribeLoadBalancers(input *elb.DescribeLoadBalancersInput) (*elb.DescribeLoadBalancersOutput, error)
	DescribeTags(input *elb.DescribeTagsInput) (*elb.DescribeTagsOutput, error)
}

// MyELBV2API is a subset of the AWS ELBV2 API interface
type MyELBV2API interface {
	DeregisterTargets(input *elbv2.DeregisterTargetsInput) (*elbv2.DeregisterTargetsOutput, error)
	DescribeListeners(input *elbv2.DescribeListenersInput) (*elbv2.DescribeListenersOutput, error)
	DescribeLoadBalancers(input *elbv2.DescribeLoadBalancersInput) (*elbv2.DescribeLoadBalancersOutput, error)
	DescribeTags(input *elbv2.DescribeTagsInput) (*elbv2.DescribeTagsOutput, error)
	DescribeTargetHealth(input *elbv2.DescribeTargetHealthInput) (*elbv2.DescribeTargetHealthOutput, error)
}

// CloudProvider is a wrapper around the required interfaces
// to allow for mocking. It implements the CloudProvider interface
type CloudProvider struct {
	EC2   MyEC2API
	ELB   MyELBAPI
	ELBV2 MyELBV2API
}

// DrainNodeFromLoadBalancer drains the node from both ELB and ELBV2 load balancers in AWS land
func (m *CloudProvider) DrainNodeFromLoadBalancer(nodeName string) error {
	log.Info().
		Str("nodeName", nodeName).
		Msg("handling deregistration for node")
	nodeID := nodeName
	if !strings.HasPrefix(nodeName, "i-") {
		// get node ID from hostname
		nodeIDFromHostname, err := m.getNodeIDFromIP(nodeName)
		if err != nil {
			return err
		}

		nodeID = *nodeIDFromHostname
	}

	vpcID, clusterName, err := m.GetVPCAndClusterFromInstance(nodeID)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(2)
	log.Info().Msg("beginning drain operations")
	go func() {
		err := m.drainNodeFromELBV1sInCluster(nodeID, *vpcID, *clusterName)
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
		err := m.drainNodeFromELBV2sInCluster(nodeID, *vpcID, *clusterName)
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

func (m *CloudProvider) drainNodeFromELBV1sInCluster(nodeID string, vpcID string, clusterName string) error {
	elbV1Names, err := m.getELBV1s(vpcID, clusterName)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, elbV1Name := range elbV1Names {
		wg.Add(1)
		start := time.Now()
		go func(name string) {
			for {
				log.Debug().
					Str("elbName", name).
					Str("nodeID", nodeID).
					Msg("draining node from ELB v1")

				drained, _ := m.drainNodeFromELBV1(nodeID, name)

				if drained {
					wg.Done()
					break
				}

				if time.Since(start) > (120 * time.Second) {
					log.Warn().
						Str("elbName", name).
						Str("nodeID", nodeID).
						Msg("node did not drain within 120s")
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

func (m *CloudProvider) drainNodeFromELBV2sInCluster(nodeID string, vpcID string, clusterName string) error {
	targetGroupARNs, err := m.getELBV2TargetGroupARNsInCluster(vpcID, clusterName)
	if err != nil {
		return nil
	}

	var wg sync.WaitGroup
	for _, targetGroupARN := range targetGroupARNs {
		wg.Add(1)
		start := time.Now()
		go func(arn string) {
			for {
				drained, _ := m.drainNodeFromELBV2TargetGroup(nodeID, arn)
				log.Debug().
					Str("elbArn", arn).
					Str("nodeID", nodeID).
					Msg("draining node from ELB v2")

				if drained {
					wg.Done()
					break
				}

				if time.Since(start) > (120 * time.Second) {
					log.Warn().
						Str("elbArn", arn).
						Str("nodeID", nodeID).
						Msg("node did not drain within 120s")
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
