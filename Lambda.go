package main

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"strconv"
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

	return resp
}

func _listLambdaFunctions() *lambda.ListFunctionsOutput {
	svc := lambda.New(session.New(), _getAwsConfig())

	params := &lambda.ListFunctionsInput{

	}

	resp, err := svc.ListFunctions(params)

	if (err != nil) {
		errState(err.Error())
	}

	return resp
}

func _deployLambdaFunction(functionName, bucket, filename, alias, version string, publish bool) {
	svc := lambda.New(session.New(), _getAwsConfig())

	params := &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(functionName),
		S3Bucket: aws.String(bucket),
		S3Key: aws.String(filename),
	}

	result, err := svc.UpdateFunctionCode(params)

	if err != nil {
		errState(err.Error())
	}

	if publish {

		params := &lambda.PublishVersionInput{
			FunctionName: aws.String(functionName),
			CodeSha256: aws.String(*result.CodeSha256),
			Description: aws.String(version),
		}

		_, errP := svc.PublishVersion(params);

		if errP != nil {
			errState(errP.Error())
		}

		paramsU := &lambda.UpdateAliasInput{
			FunctionName: aws.String(functionName),
			FunctionVersion: aws.String(*result.Version),
			Name: aws.String(alias),
		}

		_, errU := svc.UpdateAlias(paramsU)

		if (errU != nil) {
			errState(errU.Error())
		}
	}
}

func DeployLambdaFunction(functionName, bucket, filename, alias, version, publish string) {
	doPublish, err := strconv.ParseBool(publish)

	if (err != nil) {
		doPublish = false;
	}

	_deployLambdaFunction(functionName, bucket, filename, alias, version, doPublish)
}

func GetLambdaFunctionAliasInfo(functionName, alias string) {
	aliasInfo := _getLambdaFunctionAliasInfo(functionName, alias)

	functionInfo := _getLambdaFunctionInfo(functionName, *aliasInfo.FunctionVersion)

	fmt.Println(*functionInfo.Version, ": ", *functionInfo.Description);
}

func GetLambdaFunctionInfo(functionName string) {
	fmt.Println(_getLambdaFunctionInfo(functionName, "$LATEST"))
}

func ListLambdaFunctions() {

	result := _listLambdaFunctions()

	for i := 0; i < len(result.Functions); i++ {
		fmt.Println(*result.Functions[i].FunctionName)
	}
}