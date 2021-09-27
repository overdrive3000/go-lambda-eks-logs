# go-lambda-eks-logs

This is a sample lambda function that demonstrate how to send alerts based on EKS Control Plane log patterns. In this case the lambda function is triggered any time that a Pod in EKS is terminated due to OOMKilled event, this example is useful for one-shot running Pods that are not controlled via a ReplicaSet for example Kubernetes Jobs this is because the Pod will be never restarted and we want to get notified if the reason of the Job failure is the Pod reaching its memory limit.

For this example we are not using Prometheus/Alertmanager but only AWS native solutions. So, we are going to rely on Cloudwath Log Subscription Filters in order to trigger our lambda function any time that a specific pattern is detected in our logs, in this case we are looking for an `OOMKilled` event, which looks like following entry:

```
{"kind":"Event","apiVersion":"audit.k8s.io/v1","level":"Request","auditID":"0e507148-8b16-4dfb-934b-9b05dba50974","stage":"ResponseComplete","requestURI":"/api/v1/namespaces/default/pods/stress-83-tbnq2/status","verb":"patch","user":{"username":"system:node:ip-00-00-00-00.eu-west-1.compute.internal","uid":"heptio-authenticator-aws:12345678912:AROA2PRXA4MTEFOJD26WP","groups":["system:bootstrappers","system:nodes","system:authenticated"],"extra":{"accessKeyId":["AAAAAAAAAAAAAAAA"],"arn":["arn:aws:sts::12345678912:assumed-role/eksctl-tgw-eu-west-1-nodegroup-tg-NodeInstanceRole-SP426AQ4NC1Y/i-00000000000000"],"canonicalArn":["arn:aws:iam::12345678912:role/eksctl-tgw-eu-west-1-nodegroup-tg-NodeInstanceRole-SP426AQ4NC1Y"],"sessionName":["i-00000000000000"]}},"sourceIPs":["54.75.17.171"],"userAgent":"kubelet/v1.21.2 (linux/amd64) kubernetes/729bdfc","objectRef":{"resource":"pods","namespace":"default","name":"stress-83-tbnq2","apiVersion":"v1","subresource":"status"},"responseStatus":{"metadata":{},"code":200},"requestObject":{"metadata":{"uid":"10177585-b79f-44a0-bdcc-c68ec12d012a"},"status":{"$setElementOrder/conditions":[{"type":"Initialized"},{"type":"Ready"},{"type":"ContainersReady"},{"type":"PodScheduled"}],"conditions":[{"lastTransitionTime":"2021-09-27T08:36:44Z","message":"containers with unready status: [stress]","reason":"ContainersNotReady","status":"False","type":"Ready"},{"lastTransitionTime":"2021-09-27T08:36:44Z","message":"containers with unready status: [stress]","reason":"ContainersNotReady","status":"False","type":"ContainersReady"}],"containerStatuses":[{"containerID":"docker://3d27054922714f8db6302f92e218418e7454391117a74be498fea64376b20fef","image":"12345678912.dkr.ecr.eu-west-1.amazonaws.com/stress:oom-1","imageID":"docker-pullable://12345678912.dkr.ecr.eu-west-1.amazonaws.com/stress@sha256:e3040ba32bb3a4c9f20993c63892e3c26dcfab257a0cedd73afb92b350fb7f14","lastState":{},"name":"stress","ready":false,"restartCount":0,"started":false,"state":{"terminated":{"containerID":"docker://3d27054922714f8db6302f92e218418e7454391117a74be498fea64376b20fef","exitCode":2,"finishedAt":"2021-09-27T08:36:42Z","reason":"OOMKilled","startedAt":"2021-09-27T08:36:29Z"}}}],"phase":"Failed"}},"requestReceivedTimestamp":"2021-09-27T08:36:44.014671Z","stageTimestamp":"2021-09-27T08:36:44.021638Z","annotations":{"authorization.k8s.io/decision":"allow","authorization.k8s.io/reason":""}}
```
From above entry this is the important content `{"containerID":"docker://3d27054922714f8db6302f92e218418e7454391117a74be498fea64376b20fef","exitCode":2,"finishedAt":"2021-09-27T08:36:42Z","reason":"OOMKilled","startedAt":"2021-09-27T08:36:29Z"}` in which state that this pod has been terminated because of `OOMKilled` event.

Following Cloudwatch Logs Insights query will return all log entries for `OOMKilled` events:

```
fields @timestamp, @message
| filter @logStream like /kube-apiserver-audit/
| filter ispresent(requestObject.status.containerStatuses.0.state.terminated.reason)
| filter requestObject.status.containerStatuses.0.state.terminated.reason == "OOMKilled"
| sort @timestamp desc
| limit 20
```
Take a look to the log fields, so you know what information you can get from there, in our example we are going to extract only the Pod name and its Namespace.

## Files

```bash
.
├── README.md                   <-- This instructions file
├── go-lambda-eks-logs          <-- Source code for a lambda function
│   ├── main.go                 <-- Lambda function code
├── test_app                    <-- Source code for a sample application
│   ├── Dockerfile              <-- Dockerfile for sample application
│   ├── crashit.sh              <-- Bash entrypoint for sample application
│   ├── job.yaml                <-- Kubernetes Job manifest for sample application
└── template.yaml               <-- SAM Template file
```

## Requirements

* AWS CLI already configured with Administrator permission
* [Docker installed](https://www.docker.com/community-edition)
* [Golang](https://golang.org)
* SAM CLI - [Install the SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/serverless-sam-cli-install.html)
* EKS Cluster - [Create EKS Cluster](https://docs.aws.amazon.com/eks/latest/userguide/create-cluster.html)
* EKS Control Plane logs enabled - [Enable EKS Control Plane Logs](https://docs.aws.amazon.com/eks/latest/userguide/control-plane-logs.html)

## Setup process

### Installing dependencies & building the target 

In this example we use the built-in `sam build` to automatically download all the dependencies and package our build target.   
Read more about [SAM Build here](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/sam-cli-command-reference-sam-build.html) 


## Packaging and deployment

AWS Lambda Golang runtime requires a flat folder with the executable generated on build step. SAM will use `CodeUri` property to know where to look up for the application:

```yaml
...
    FirstFunction:
        Type: AWS::Serverless::Function
        Properties:
            CodeUri: go-lambda-eks-logs/
            ...
```

To deploy your application for the first time, run the following in your shell:

```bash
sam deploy sam deploy --parameter-overrides SNSEmail=<EMAIL_TO_SEND_NOTIFICATION> ClusterName=<MY_CLUSTER> --guided
```

The command will package and deploy your application to AWS, with a series of prompts:

* **Stack Name**: The name of the stack to deploy to CloudFormation. This should be unique to your account and region, and a good starting point would be something matching your project name.
* **AWS Region**: The AWS region you want to deploy your app to.
* **Confirm changes before deploy**: If set to yes, any change sets will be shown to you before execution for manual review. If set to no, the AWS SAM CLI will automatically deploy application changes.
* **Allow SAM CLI IAM role creation**: Many AWS SAM templates, including this example, create AWS IAM roles required for the AWS Lambda function(s) included to access AWS services. By default, these are scoped down to minimum required permissions. To deploy an AWS CloudFormation stack which creates or modifies IAM roles, the `CAPABILITY_IAM` value for `capabilities` must be provided. If permission isn't provided through this prompt, to deploy this example you must explicitly pass `--capabilities CAPABILITY_IAM` to the `sam deploy` command.
* **Save arguments to samconfig.toml**: If set to yes, your choices will be saved to a configuration file inside the project, so that in the future you can just re-run `sam deploy` without parameters to deploy changes to your application.

You can find your API Gateway Endpoint URL in the output values displayed after deployment.

### Testing
To test the Cloudwatch Log Subscription Filter and Lambda function you can use the sample app provided in the `test_app` folder, this app basically will execute a container via a Kubernetes Job that will start sleep for a random period of 1 to 10 seconds and then allocate more memory than the specified in the pod resource limit.

```bash
kubectl apply -f test_app/job.ymal
```
You should be able to see the lambda execution in the AWS Lambda console in a couple of minutes and if everything goes well you should receive an email similar to:

```
Pod stress-ftss6 in namespace default has been OOMKilled
```

# Appendix

### Golang installation

Please ensure Go 1.x (where 'x' is the latest version) is installed as per the instructions on the official golang website: https://golang.org/doc/install

A quickstart way would be to use Homebrew, chocolatey or your linux package manager.

#### Homebrew (Mac)

Issue the following command from the terminal:

```shell
brew install golang
```

If it's already installed, run the following command to ensure it's the latest version:

```shell
brew update
brew upgrade golang
```

#### Chocolatey (Windows)

Issue the following command from the powershell:

```shell
choco install golang
```

If it's already installed, run the following command to ensure it's the latest version:

```shell
choco upgrade golang
```