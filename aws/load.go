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

type LoadFlag struct {
	Path string
	Prefix string
	Delimiter string
	Template string
	Region string
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
		variables = client.loadVariablesByPath(flag.Path, make(map[string]string), nil)
	} else {
		variables = client.loadVariables(flag.Prefix, make(map[string]string), nil)
	}

	values := []string{}
	for k, v := range variables {
		t, err := template.New("v").Parse(flag.Template)
		if err != nil {
			log.SetOutput(os.Stderr)
			log.Print("Template Rendering Error", err)
			os.Exit(1)
		}
		buf := &bytes.Buffer{}
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
	config := aws.Config{}
	if region != "" {
		config.Region = &region
	}
	sess := session.Must(session.NewSession(&config))
	c.Client = ssm.New(sess)
}

func (c *Client) loadVariablesByPath(path string, acc map[string]string, nextToken *string) map[string] string {

	input := &ssm.GetParametersByPathInput{
		Path: aws.String(path),
		WithDecryption: aws.Bool(true),
	}

	if nextToken != nil {
		input.SetNextToken(*nextToken)
	}
	output, err := c.Client.GetParametersByPath(input)

	if err != nil {
		log.SetOutput(os.Stderr)
		log.Print("GetParametersByPath Error", err)
		os.Exit(1)
	}
	for _, element := range output.Parameters {
		name := *element.Name
		key := strings.Trim(strings.Replace(name, path, "", 1), "/")
		acc[key] = *element.Value
	}

	if output.NextToken == nil {
		return acc
	} else {
		return c.loadVariablesByPath(path, acc, output.NextToken)
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
		log.SetOutput(os.Stderr)
		log.Print("DescribeParameters Error", err)
		os.Exit(1)
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
		log.SetOutput(os.Stderr)
		log.Print("GetParameters Error", err)
		os.Exit(1)
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