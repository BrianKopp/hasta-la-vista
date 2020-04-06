# ASG Lifecycle Hook

You can subscribe a lambda function to the ASG Termination Lifecycle hook.

## Quick Start

```bash
git clone https://github.com/briankopp/hasta-la-vista
cd cmd/lambda-asg-term-hook
sh ./build.sh
ls -al out # build.sh copies lambda-compatible zip artifact
```

## CloudWatch Events

Messages coming through CloudWatch look like:

```json
{
  "version": "0",
  "id": "12345678-1234-1234-1234-123456789012",
  "detail-type": "EC2 Instance-terminate Lifecycle Action",
  "source": "aws.autoscaling",
  "account": "123456789012",
  "time": "yyyy-mm-ddThh:mm:ssZ",
  "region": "us-west-2",
  "resources": [
    "auto-scaling-group-arn"
  ],
  "detail": {
    "LifecycleActionToken":"87654321-4321-4321-4321-210987654321",
    "AutoScalingGroupName":"my-asg",
    "LifecycleHookName":"my-lifecycle-hook",
    "EC2InstanceId":"i-1234567890abcdef0",
    "LifecycleTransition":"autoscaling:EC2_INSTANCE_TERMINATING",
    "NotificationMetadata":"additional-info"
  }
}```

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
