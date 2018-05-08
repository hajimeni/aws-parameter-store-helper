package client

import (
	"fmt"
	"strings"
	"text/template"
	"log"
	"bytes"
	"strconv"
	"github.com/kballard/go-shellquote"
)

type LoadFlag struct {
	Path            []string
	Prefix          []string
	Delimiter       string
	Template        string
	Region          string
	Recursive       bool
	UpperCaseKey    bool
	ReplaceKeys     string
	ReplaceKeyValue string
	QuoteShell      bool
    NoQuoteShell    bool
}

type Loader struct {
	Client Client
}

func LoadParameterStore(client Client, flag *LoadFlag) {
	var variables []KeyValue
	if len(flag.Path) > 0 {
		variables = append(variables, client.LoadVariablesByPaths(flag.Path, flag.Recursive)...)
	}
	if len(flag.Prefix) > 0 {
		variables = append(variables, client.LoadVariablesByPrefixes(flag.Prefix)...)
	}

	fmt.Print(renderTemplate(&variables, flag))
}

func renderTemplate(variables *[]KeyValue, flag *LoadFlag) string {
	krs := []string{}
	for _, rv := range strings.Split(flag.ReplaceKeys, "") {
		krs = append(krs, rv, flag.ReplaceKeyValue)
	}
	kr := strings.NewReplacer(krs...)

	t, err := template.New("v").Parse(flag.Template)
	if err != nil {
		log.Fatal("Template Rendering Error", err)
	}
	values := []string{}
	for _, kv := range *variables {
		buf := &bytes.Buffer{}

		k := kr.Replace(kv.Key)
		v := kv.Value
		if flag.UpperCaseKey {
			k = strings.ToUpper(k)
		}
		if flag.QuoteShell && !flag.NoQuoteShell {
			v = shellquote.Join(v)
		}
		t.Execute(buf, map[string]string{"Name": k, "Value": v})
		values = append(values, buf.String())
	}
	s, err := strconv.Unquote("\"" + flag.Delimiter + "\"")
	if err != nil {
		s = flag.Delimiter
	}
	return strings.Join(values, s)
}
