# AWS Cloud Provider

In order to run this function using AWS as the cloud provider, pass
the environment variable `CLOUDPROVIDER=aws`. You'll need to
specify the `AWS_REGION` environment variable and bring your
own credentials, e.g. by using the lambda execution role.

## How It Works

The `hasta-la-vista` deregistration process follows the following
steps:

* Acquire the EC2 Instance ID if not already present
(e.g. if node passed in is the node private
DNS name, lookup the instance ID).
* Determine the VPC ID and cluster name that the instance is in.
  * Expects instance to have tag with key `kubernetes.io/<cluster_name>/owned`.
* Find all ELBs (classic load balancers) and v2 ELBs
(network/application load balancers) that are in that VPC
and have a tag with key `kubernetes.io/<cluster_name>/owned`.
* For each of the ELBs, check if instance is in service. If so,
deregister the instance from the ELB and check until TIMEOUT
for it to no longer be registered to that ELB.
* For each of the v2 ELBs, find all target groups.
* For each of the target groups, check if instance is in service,
if so, deregister it. Wait until instance is no longer a `healthy`
or `draining` member of that target group.
