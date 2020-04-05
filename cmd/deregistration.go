package main

import (
	"errors"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/rs/zerolog/log"
)

var awsSession = session.Must(session.NewSession())

func (m *handler) handleDeregistration(nIP string, nID string, clusterName string, vpcID string) error {
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
		id, err := m.getNodeIDFromIP(nIP)
		if err != nil {
			return err
		}

		log.Info().Str("nodeID", *id).Msg("acquired node id from node ip")
		nodeID = *id
	}

	var wg sync.WaitGroup
	log.Info().Msg("beginning drain operations")
	go func() {
		err := m.drainNodeFromELBV1sInCluster(nodeID, clusterName, vpcID)
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
		err := m.drainNodeFromELBV2sInCluster(nodeID, clusterName, vpcID)
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

func (m *handler) drainNodeFromELBV1sInCluster(nodeID string, clusterName string, vpcID string) error {
	elbV1Names, err := m.getELBV1s(clusterName, vpcID)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, elbV1Name := range elbV1Names {
		wg.Add(1)
		start := time.Now()
		go func(name string) {
			for {
				drained, _ := m.drainNodeFromELBV1(nodeID, name)
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

func (m *handler) drainNodeFromELBV2sInCluster(nodeID string, clusterName string, vpcID string) error {
	targetGroupARNs, err := m.getELBV2TargetGroupARNsInCluster(clusterName, vpcID)
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
