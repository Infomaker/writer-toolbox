package main

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
)

func _getLambdaFunctionAliasInfo(functionName, alias string) *lambda.AliasConfiguration {
	svc := lambda.New(session.New(), _getAwsConfig())

	params := &lambda.GetAliasInput{
		FunctionName: aws.String(functionName),
		Name: aws.String(alias),
	}

	resp, err := svc.GetAlias(params);

	if err != nil {
		errState(err.Error())
	}

	return resp;
}

func _getLambdaFunctionInfo(functionName, qualifier string) *lambda.FunctionConfiguration {
	svc := lambda.New(session.New(), _getAwsConfig())

	params := &lambda.GetFunctionConfigurationInput{
		FunctionName: aws.String(functionName),
		Qualifier: aws.String(qualifier),
	}

	resp, err := svc.GetFunctionConfiguration(params);

	if err != nil {
		errState(err.Error())
	}

	return resp;
}

func GetLambdaFunctionAliasInfo(functionName, alias string) {
	aliasInfo := _getLambdaFunctionAliasInfo(functionName, alias)

	functionInfo := _getLambdaFunctionInfo(functionName, *aliasInfo.FunctionVersion)

	fmt.Println(*functionInfo.Description);
}

func GetLambdaFunctionInfo(functionName string) {
	fmt.Println(_getLambdaFunctionInfo(functionName, "$LATEST"))
}