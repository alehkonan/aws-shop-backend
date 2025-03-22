package authorization

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewStack(scope constructs.Construct, id string, props *awscdk.StackProps) awscdk.Stack {
	stack := awscdk.NewStack(scope, &id, props)

	basicAuthorizerFunction := awslambda.NewFunction(
		stack, jsii.String("BasicAuthorizerFunction"),
		&awslambda.FunctionProps{
			Runtime: awslambda.Runtime_PROVIDED_AL2023(),
			Code:    awslambda.Code_FromAsset(jsii.String("lambdas/basicAuthorizer"), nil),
			Handler: jsii.String("bootstrap"),
			Environment: &map[string]*string{
				"USERNAME": jsii.String(os.Getenv("USERNAME")),
				"PASSWORD": jsii.String(os.Getenv("PASSWORD")),
			},
		},
	)

	awscdk.NewCfnOutput(stack, jsii.String("BasicAuthorizerFunctionArn"),
		&awscdk.CfnOutputProps{
			Value:      basicAuthorizerFunction.FunctionArn(),
			ExportName: jsii.String("BasicAuthorizerFunctionArn"),
		},
	)

	return stack
}
