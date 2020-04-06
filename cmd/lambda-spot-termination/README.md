# Spot Termination

You can subscribe a lambda function to the spot termination handler.

Messages coming through CloudWatch look like:

```json
{
  "version": "0",
  "id": "54f8f720-c5ab-1de8-721f-648455223dbf",
  "detail-type": "EC2 Spot Instance Interruption Warning",
  "source": "aws.ec2",
  "account": "274122010097",
  "time": "2020-03-27T08:50:06Z",
  "region": "us-east-1",
  "resources": [ "arn:aws:ec2:us-east-1a:instance/i-04674c0a5f28fa9e5" ],
  "detail": {
    "instance-id": "i-04674c0a5f28fa9e5",
    "instance-action": "terminate"
  }
}
```

