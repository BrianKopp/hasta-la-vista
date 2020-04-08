# Hasta-La-Vista
![Go Build](https://github.com/BrianKopp/hasta-la-vista/workflows/Go/badge.svg)

A simple golang application to ensure that nodes are fully deregistered
from Elastic Load Balancers.

## Usage

```bash
curl -X POST http://<hostname>/drain?id=i-abcdefg&pw=api-key
```

Provide either the instance `id` or `ip` of the instance.
Include a secret in the `pw`.

## Requirements

### IAM

The application requires the following IAM policy.

```json
TODO
```

### AWS Configuration

It is expected that your cluster is in a single VPC.
Only those load balancers in the provided VPC will be checked.

Your load balancers are expected to have a tag
with the key: `kubernetets/cluster/<cluster_name>`

### Environment Variables

| Name | Description | Default |
|:----:|:----------- |:-------:|
|SECRET|the secret with which to protect your API|N/A|
|LOGLEVEL|the logging verbosity, accepts `debug`, `info`, `warn` and `error`|`info`|
|CLOUDPROVIDER|the type of cloud provider, options (`aws`)|N/A|
|TIMEOUT|the max amount of time the function will wait for the node to deregister|N\A|
|DRYRUN|whether to operate in a "dry run" mode. No write actions are performed|`false`|
|AWS_REGION|the AWS region you're in|N/A|

## Problem Cases

There are several kubernetes events that benefit from manually
ensuring node deregistration.

### Instance Scale Down

Typically, `LoadBalancer` services are configured to
allow traffic from an ELB to any node in the cluster.
The receiving node checks to see if a pod for that service
is local on the node. If not, it proxies the request to
another node in the cluster with a pod for that service.

When a node is drained and preparing for termination,
it may still be proxying requests. If the ELB health
check has not fully failed by the time the node
gets terminated, you could have a case where the
ELB is still attempting to send traffic to a node that is
not capable of handling it.

### Spot Termination

This situation is just like the scale down situation,
except that the time frame is shorter. Your node
may not have time to fail out of ELB health checks
and gracefully drain by the time it is violently terminated.
