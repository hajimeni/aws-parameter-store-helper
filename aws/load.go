package aws

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"log"
	"bytes"
	"strconv"
	"os"
	"github.com/pkg/errors"
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

type LoadFlag struct {
	Path string
	Prefix string
	Delimiter string
	Template string
	Region string
	Recursive bool
	UpperCaseKey bool
	ReplaceKeys string
	ReplaceKeyValue string
	EscapeDoublequote string
}

type Client struct {
	Client *ssm.SSM
}

var client Client = Client{}

func CheckRequiredFlags(flag *LoadFlag) error {
	if (flag.Path == "" && flag.Prefix == "") {
		return errors.New("Required Path or Prefix")
	}
	client.createClient(flag.Region)

	return nil
}

func LoadParameterStore(flag *LoadFlag) {
	var variables map[string]string
	if flag.Path != "" {
		variables = client.loadVariablesByPath(flag.Path, flag.Recursive, make(map[string]string), nil)
	} else {
		variables = client.loadVariables(flag.Prefix, make(map[string]string), nil)
	}

	krs := []string{}
	for _, rv := range strings.Split(flag.ReplaceKeys, "") {
		krs = append(krs, rv, flag.ReplaceKeyValue)
	}
	kr := strings.NewReplacer(krs...)

	values := []string{}
	for k, v := range variables {
		t, err := template.New("v").Parse(flag.Template)
		if err != nil {
			log.Fatal("Template Rendering Error", err)
		}
		buf := &bytes.Buffer{}

		k = kr.Replace(k)
		if flag.UpperCaseKey {
			k = strings.ToUpper(k)
		}
		if flag.EscapeDoublequote != "" {
			v = strings.Replace(v, "\"", flag.EscapeDoublequote + "\"", -1)
		}
		t.Execute(buf, map[string]string{"Name": k, "Value": v})
		values = append(values, buf.String())
	}
	s, err := strconv.Unquote("\"" + flag.Delimiter + "\"")
	if err != nil {
		s = flag.Delimiter
	}
	fmt.Print(strings.Join(values, s))
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

func (c *Client) loadVariablesByPath(path string, recursive bool, acc map[string]string, nextToken *string) map[string] string {

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
		acc[key] = *element.Value
	}

	if output.NextToken == nil {
		return acc
	} else {
		return c.loadVariablesByPath(path, recursive, acc, output.NextToken)
	}
}


func (c Client) loadVariables(prefix string, acc map[string]string, nextToken *string) map[string] string {

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
			acc[key] = *element.Value
		}
	}

	if output.NextToken == nil {
		return acc
	} else {
		return c.loadVariables(prefix, acc, output.NextToken)
	}
}