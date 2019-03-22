package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"strconv"
)

func DeployLambdaFunction(functionName, bucket, filename, alias, version, runtime, publish string) {
	doPublish, err := strconv.ParseBool(publish)

	if err != nil {
		doPublish = false
	}

	deployLambdaFunction(functionName, bucket, filename, alias, version, runtime, doPublish)
}

func GetLambdaFunctionAliasInfo(functionName, alias string) {
	aliasInfo := getLambdaFunctionAliasInfo(functionName, alias)
	functionInfo := getLambdaFunctionInfo(functionName, *aliasInfo.FunctionVersion)

	fmt.Println(*functionInfo.Version, ": ", *functionInfo.Description)
}

func GetLambdaFunctionInfo(functionName string) {
	fmt.Println(getLambdaFunctionInfo(functionName, "$LATEST"))
}

func ListLambdaFunctions() {
	result := listLambdaFunctions()

	for i := 0; i < len(result.Functions); i++ {
		fmt.Println(*result.Functions[i].FunctionName)
	}
}

func deployLambdaFunction(functionName, bucket, filename, alias, version, runtime string, publish bool) {
	sess, cfg := getSessionAndConfig()
	svc := lambda.New(sess, cfg)

	if runtime != "" {
		params := &lambda.UpdateFunctionConfigurationInput{
			FunctionName: aws.String(functionName),
			Runtime:      aws.String(runtime),
		}

		result, err := svc.UpdateFunctionConfiguration(params)
		assertError(err)

		fmt.Printf("Updated function %s with configuration %s\n", *result.FunctionName, *result.Runtime)
	}

	params := &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(functionName),
		S3Bucket:     aws.String(bucket),
		S3Key:        aws.String(filename),
	}

	result, err := svc.UpdateFunctionCode(params)
	assertError(err)

	fmt.Printf("Updated %s with shasum %s\n", *result.FunctionName, *result.CodeSha256)

	if publish {
		params := &lambda.PublishVersionInput{
			FunctionName: aws.String(functionName),
			CodeSha256:   aws.String(*result.CodeSha256),
			Description:  aws.String(version),
		}

		published, errP := svc.PublishVersion(params)

		if errP != nil {
			errState(errP.Error())
		}

		paramsU := &lambda.UpdateAliasInput{
			FunctionName:    aws.String(functionName),
			FunctionVersion: aws.String(*published.Version),
			Name:            aws.String(alias),
		}

		_, errU := svc.UpdateAlias(paramsU)

		if errU != nil {
			errState(errU.Error())
		}

		fmt.Printf("Alias %s updated to point to version %s (%s)\n", alias, *published.Version, *published.Description)
	}
}

func getLambdaFunctionInfo(functionName, qualifier string) *lambda.FunctionConfiguration {
	sess, cfg := getSessionAndConfig()
	svc := lambda.New(sess, cfg)

	params := &lambda.GetFunctionConfigurationInput{
		FunctionName: aws.String(functionName),
		Qualifier:    aws.String(qualifier),
	}

	resp, err := svc.GetFunctionConfiguration(params)
	assertError(err)

	return resp
}

func listLambdaFunctions() *lambda.ListFunctionsOutput {
	sess, cfg := getSessionAndConfig()
	svc := lambda.New(sess, cfg)

	var marker = new(string)
	var result = new(lambda.ListFunctionsOutput)

	for marker != nil && len(result.Functions) < int(maxResult) {
		if marker != nil && *marker == "" {
			marker = nil
		}

		params := &lambda.ListFunctionsInput{
			Marker:   marker,
			MaxItems: &maxResult,
		}

		resp, err := svc.ListFunctions(params)
		assertError(err)

		result.Functions = append(result.Functions, resp.Functions...)
		marker = resp.NextMarker
	}

	return result
}

func getLambdaFunctionAliasInfo(functionName, alias string) *lambda.AliasConfiguration {
	sess, cfg := getSessionAndConfig()
	svc := lambda.New(sess, cfg)

	params := &lambda.GetAliasInput{
		FunctionName: aws.String(functionName),
		Name:         aws.String(alias),
	}

	resp, err := svc.GetAlias(params)
	assertError(err)

	return resp
}
