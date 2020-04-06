# Spot Termination

You can subscribe a lambda function to the spot termination handler.

## Quick Start

```bash
git clone https://github.com/briankopp/hasta-la-vista
cd cmd/lambda-spot-termination
sh ./build.sh
ls -al out # build.sh copies lambda-compatible zip artifact
```

## CloudWatch Events

Messages coming through CloudWatch look like:

```json
{
  "version": "0",
  "id": "84f8f720-c5ab-1de8-721f-648455223dbf",
  "detail-type": "EC2 Spot Instance Interruption Warning",
  "source": "aws.ec2",
  "account": "0123456789",
  "time": "2020-01-01T01:00:00Z",
  "region": "us-east-1",
  "resources": [ "arn:aws:ec2:us-east-1a:instance/i-0123456789" ],
  "detail": {
    "instance-id": "i-0123456789",
    "instance-action": "terminate"
  }
}
```

## IAM Permissions

The lambda function role needs to have the following policy.

```json
[
  {
    "Sid": "ReadOnlyPermissions",
    "Effect": "Allow",
    "Action": [
      "ec2:DescribeInstances"
      "elb:DescribeInstanceHealth",
      "elb:DescribeLoadBalancers",
      "elb:DescribeTags",
      "elbv2:DescribeListeners",
      "elbv2:DescribeLoadBalancers",
      "elbv2:DescribeTags",
      "elbv2:DescribeTargetHealth"
    ],
    "Resources": "*"
  }, {
    "Sid": "DeregisterPermissionForELBv1s",
    "Effect": "Allow",
    "Action": [
      "elb:DeregisterInstancesFromLoadBalancer"
    ],
    "Resources": "*" // restrict accordingly
  }, {
    "Sid": "DeregisterPermissionForELBv1s",
    "Effect": "Allow",
    "Action": [
      "elbv2:DeregisterTargets"
    ],
    "Resources": "*" // restrict accordingly
  }
]
```

The role should also have the required permissions to write CW Logs.
