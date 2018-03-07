package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/aws/session"
	"log"
	"os"
	"strings"
)

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

type Client struct {
	Client *ssm.SSM
}

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

func (c *Client) createClient(region string) {
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
	c.Client = ssm.New(sess, config)
}

func (c *Client) loadVariablesByPaths(paths *[]string, recursive bool) []KeyValue {
	res := []KeyValue{}
	for _, path := range *paths {
		r := c.loadVariablesByPath(path, recursive, res, nil)
		for k, v := range r {
			res[k] = v
		}
	}
	return res
}

func (c *Client) loadVariablesByPath(path string, recursive bool, acc []KeyValue, nextToken *string) []KeyValue {

	input := &ssm.GetParametersByPathInput{
		Path: aws.String(path),
		WithDecryption: aws.Bool(true),
	}
	input.Recursive = aws.Bool(recursive)

	if nextToken != nil {
		input.SetNextToken(*nextToken)
	}
	output, err := c.Client.GetParametersByPath(input)

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


func (c *Client) loadVariablesByPrefixes(prefixes *[]string) []KeyValue {
	res := []KeyValue{}
	for _, prefix := range *prefixes {
		r := c.loadVariables(prefix, res, nil)
		for k, v := range r {
			res[k] = v
		}
	}
	return res
}

func (c *Client) loadVariables(prefix string, acc []KeyValue, nextToken *string) []KeyValue {

	input := &ssm.DescribeParametersInput{
		MaxResults: aws.Int64(10),
	}
	if nextToken != nil {
		input.SetNextToken(*nextToken)
	}
	output, err := c.Client.DescribeParameters(input)

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
	poutput, err := c.Client.GetParameters(pintput)
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