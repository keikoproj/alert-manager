# AWS Integration for Alert Manager

This guide explains how to set up alert-manager with AWS integration, particularly using IAM Roles for Service Accounts (IRSA) on Amazon EKS clusters.

## Overview

When running alert-manager on Amazon EKS, you can leverage IAM Roles for Service Accounts (IRSA) to:

1. Securely provide AWS permissions to the alert-manager controller
2. Avoid storing long-lived AWS credentials
3. Follow the principle of least privilege

This is especially useful if your alert-manager implementation needs to interact with AWS services, such as:
- Reading tags from EC2 instances for alert enrichment
- Sending alerts to SNS topics
- Accessing metrics in CloudWatch
- Retrieving configuration from SSM Parameter Store or Secrets Manager

## Prerequisites

- An EKS cluster with IRSA enabled (has an OIDC provider)
- AWS CLI configured with administrative permissions
- kubectl configured to work with your EKS cluster
- [eksctl](https://eksctl.io/) (optional, for easier setup)

## Setting Up IRSA for Alert Manager

### 1. Verify OIDC Provider Configuration

First, check if your EKS cluster has an OIDC provider configured:

```bash
aws eks describe-cluster --name your-cluster-name --query "cluster.identity.oidc.issuer" --output text
```

If this command returns a URL, your cluster has an OIDC provider configured. If it returns no output, you need to set up the OIDC provider:

```bash
eksctl utils associate-iam-oidc-provider --cluster your-cluster-name --approve
```

### 2. Create an IAM Policy

Create an IAM policy that grants only the permissions needed by alert-manager. Here's an example policy for basic AWS service interactions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeInstances",
        "ec2:DescribeTags",
        "cloudwatch:GetMetricData",
        "sns:Publish"
      ],
      "Resource": "*"
    }
  ]
}
```

Save this as `alert-manager-policy.json` and create the policy:

```bash
aws iam create-policy \
  --policy-name AlertManagerPolicy \
  --policy-document file://alert-manager-policy.json
```

Note the ARN of the policy, as you'll need it in the next step.

### 3. Create a Service Account with IAM Role

Create a Kubernetes service account for alert-manager that's associated with an IAM role:

```bash
eksctl create iamserviceaccount \
  --name alert-manager-controller \
  --namespace alert-manager-system \
  --cluster your-cluster-name \
  --attach-policy-arn arn:aws:iam::YOUR_ACCOUNT_ID:policy/AlertManagerPolicy \
  --approve
```

Alternatively, you can use iam-manager as described [here](https://github.com/keikoproj/iam-manager) to create the necessary IAM role.

### 4. Update the Alert Manager Deployment

If you're installing alert-manager using the Makefile, update the deployment to use the new service account:

```bash
# Edit the deployment manifest
vim config/manager/manager.yaml

# Add or modify the serviceAccountName field
# serviceAccountName: alert-manager-controller
```

If alert-manager is already deployed, patch the deployment:

```bash
kubectl patch deployment alert-manager-controller-manager \
  -n alert-manager-system \
  -p '{"spec":{"template":{"spec":{"serviceAccountName":"alert-manager-controller"}}}}'
```

### 5. Verify the Configuration

Verify that the service account has the AWS role annotation:

```bash
kubectl get serviceaccount alert-manager-controller -n alert-manager-system -o yaml
```

You should see an annotation like:
```
annotations:
  eks.amazonaws.com/role-arn: arn:aws:iam::YOUR_ACCOUNT_ID:role/eks-alert-manager-role
```

## Using AWS Services in Alert Manager

Once IRSA is set up, alert-manager can use the AWS SDK to access AWS services without explicit credentials. The AWS SDK will automatically use the IAM role associated with the service account.

### Example: Accessing EC2 Instance Tags

If your alert-manager needs to access EC2 instance tags to enrich alerts, you can use the AWS SDK in your controller code:

```go
import (
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/ec2"
)

func getEC2Tags(instanceID string) (map[string]string, error) {
    // Session will use IRSA credentials automatically
    sess := session.Must(session.NewSession())
    ec2Svc := ec2.New(sess)
    
    input := &ec2.DescribeTagsInput{
        Filters: []*ec2.Filter{
            {
                Name:   aws.String("resource-id"),
                Values: []*string{aws.String(instanceID)},
            },
        },
    }
    
    result, err := ec2Svc.DescribeTags(input)
    if err != nil {
        return nil, err
    }
    
    tags := make(map[string]string)
    for _, tag := range result.Tags {
        tags[*tag.Key] = *tag.Value
    }
    
    return tags, nil
}
```

## Troubleshooting IRSA

If you're experiencing issues with IRSA:

1. **Check the role trust policy** - Make sure the role trust policy correctly references your EKS cluster's OIDC provider and the service account.

2. **Verify pod identity** - You can check if the pod is correctly assuming the IAM role by running:
   ```bash
   kubectl exec -it -n alert-manager-system deploy/alert-manager-controller-manager -- aws sts get-caller-identity
   ```

3. **Check controller logs** - Look for AWS-related errors in the controller logs:
   ```bash
   kubectl logs -n alert-manager-system deploy/alert-manager-controller-manager
   ```

4. **Use AWS SDK debug logging** - Enable AWS SDK debug logging by setting the environment variable:
   ```
   AWS_SDK_GO_LOG_LEVEL=Debug
   ```

## Security Considerations

- Follow the principle of least privilege - only grant permissions that are actually needed
- Regularly audit and rotate any long-lived credentials
- Consider using condition keys in your IAM policies to further restrict access
- Use aws:PrincipalTag conditions to ensure the role can only be assumed by specific pods

## References

- [AWS IAM Roles for Service Accounts](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html)
- [AWS SDK for Go Documentation](https://docs.aws.amazon.com/sdk-for-go/api/)
- [EKS Workshop IRSA Guide](https://www.eksworkshop.com/beginner/110_irsa/)
