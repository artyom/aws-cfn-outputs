Command aws-cfn-outputs prints AWS CloudFormation [stack outputs],
optionally substituting values in a text template.

Consider that you have a CloudFormation template with the following outputs section:

```yaml
Outputs:
  StackVPC:
    Description: The ID of the VPC
    Value: !Ref MyVPC
    Export:
      Name: !Sub "${AWS::StackName}-VPCID"
```

In a template you can refer to the “StackVPC” value like so:

```
Here's my VPC ID: {{.StackVPC}}
```

To render the template, call the command as:

```
aws-cfn-outputs -s stack-name -t template.txt
```

For the full template syntax see Go standard library [text/template] package.

[stack outputs]: https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/outputs-section-structure.html
[text/template]: https://pkg.go.dev/text/template
