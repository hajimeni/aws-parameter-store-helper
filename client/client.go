package client

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/aws/session"
	"log"
	"os"
	"strings"
)

type KeyValue struct {
	Key string
	Value string
}

func newKeyValue(key string, value string) KeyValue {
	return KeyValue{
		key,
		value,
	}
}

type Client interface {
	LoadVariablesByPaths(paths []string, recursive bool) []KeyValue
	LoadVariablesByPrefixes(prefixes []string) []KeyValue
}


type AwsSsmClient struct {
	client *ssm.SSM
}

type stdErrLogger struct {
	logger *log.Logger
}

func newStdErrLogger() aws.Logger {
	return &stdErrLogger{
		logger: log.New(os.Stderr, "", log.LstdFlags),
	}
}

func (l stdErrLogger) Log(args ...interface{}) {
	l.logger.Println(args...)
}

func NewClient(region string) (Client, error) {
	log.SetOutput(os.Stderr)

	sess := session.Must(session.NewSession())
	logLevel := sess.Config.LogLevel
	if os.Getenv("DEBUG") != "" {
		logLevel = aws.LogLevel(aws.LogDebug)
	}
	if os.Getenv("DEBUG_SIGNING") != "" {
		logLevel = aws.LogLevel(aws.LogDebugWithSigning)
	}
	if os.Getenv("DEBUG_BODY") != "" {
		logLevel = aws.LogLevel(aws.LogDebugWithSigning | aws.LogDebugWithHTTPBody)
	}

	config := aws.NewConfig().WithLogger(newStdErrLogger()).WithLogLevel(*logLevel)
	if region != "" {
		config = config.WithRegion(region)
	}
	c := AwsSsmClient{
		ssm.New(sess, config),
	}

	return c, nil
}

func (c AwsSsmClient) LoadVariablesByPaths(paths []string, recursive bool) []KeyValue {
	res := []KeyValue{}
	for _, path := range paths {
		r := c.loadVariablesByPath(path, recursive, []KeyValue{}, nil)
		res = append(res, r...)
	}
	return res
}

func (c AwsSsmClient) loadVariablesByPath(path string, recursive bool, acc []KeyValue, nextToken *string) []KeyValue {
	input := &ssm.GetParametersByPathInput{
		Path: aws.String(path),
		WithDecryption: aws.Bool(true),
		Recursive: aws.Bool(recursive),
	}

	if nextToken != nil {
		input.SetNextToken(*nextToken)
	}
	output, err := c.client.GetParametersByPath(input)

	if err != nil {
		log.Fatal("GetParametersByPath Error:\n", err)
	}
	for _, element := range output.Parameters {
		name := *element.Name
		key := strings.Trim(strings.Replace(name, path, "", 1), "/")
		acc = append(acc, newKeyValue(key, *element.Value))
	}

	if output.NextToken == nil {
		return acc
	} else {
		return c.loadVariablesByPath(path, recursive, acc, output.NextToken)
	}
}

func (c AwsSsmClient) LoadVariablesByPrefixes(prefixes []string) []KeyValue {
	res := []KeyValue{}
	for _, prefix := range prefixes {
		r := c.loadVariables(prefix, res, nil)
		res = append(res, r...)
	}
	return res
}

func (c AwsSsmClient) loadVariables(prefix string, acc []KeyValue, nextToken *string) []KeyValue {
	input := &ssm.DescribeParametersInput{
		MaxResults: aws.Int64(10),
	}
	if nextToken != nil {
		input.SetNextToken(*nextToken)
	}
	output, err := c.client.DescribeParameters(input)

	if err != nil {
		log.Fatal("DescribeParameters Error", err)
	}
	names := []*string{}
	for _, v := range output.Parameters {
		names = append(names, v.Name)
	}
	pintput := &ssm.GetParametersInput{
		Names: names,
		WithDecryption: aws.Bool(true),
	}
	poutput, err := c.client.GetParameters(pintput)
	if err != nil {
		log.Fatal("GetParameters Error", err)
	}
	for _, element := range poutput.Parameters {
		name := *element.Name
		if strings.Index(name, prefix) == 0 {
			key := strings.Replace(name, prefix, "",1)
			acc = append(acc, newKeyValue(key, *element.Value))
		}
	}

	if output.NextToken == nil {
		return acc
	} else {
		return c.loadVariables(prefix, acc, output.NextToken)
	}
}