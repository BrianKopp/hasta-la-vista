# Hasta-La-Vista

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
| SECRET | the secret with which to protect your API | N/A |
| AWS_REGION | the AWS region you're in | N/A |
| LOGLEVEL | the logging verbosity, accepts `debug`, `info`, `warn` and `error` | `info` |
| CLUSTERNAME | your cluster`s name, used for tag lookups | N/A |
| VPCID | the VPC ID of your cluster | N/A |
| CLOUDPROVIDER | the type of cloud provider, options (`aws`) | N/A |

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

### Proxy Services

Kubernetes services with `externalTrafficPolicy: Local`
are notoriously challenging to work with. Nodes are
configured to fail ELB health checks if they don't
have a pod for the service on them. Nodes that do
have the service pass health checks. Thus, the ELB
only sends traffic to the nodes with the pods on them.

It also means that when a pod gets a `SIGTERM`,
it moves into a `Terminating` state, where it no longer
receives traffic. There will be a period of time where the
node hasn't failed its ELB health checks, but is unable
to handle traffic.
