package aws

import (
	"fmt"
	"strings"
	"text/template"
	"log"
	"bytes"
	"strconv"
	"github.com/pkg/errors"
)

type LoadFlag struct {
	Path []string
	Prefix []string
	Delimiter string
	Template string
	Region string
	Recursive bool
	UpperCaseKey bool
	ReplaceKeys string
	ReplaceKeyValue string
	EscapeDoublequote string
}

var client Client = Client{}

func CheckRequiredFlags(flag *LoadFlag) error {
	if (len(flag.Path) == 0 && len(flag.Prefix) == 0) {
		return errors.New("Required Path or Prefix")
	}
	client.createClient(flag.Region)

	return nil
}

func LoadParameterStore(flag *LoadFlag) {
	var variables []KeyValue
	if len(flag.Path) > 0 {
		variables = client.loadVariablesByPaths(&flag.Path, flag.Recursive)
	} else {
		variables = client.loadVariablesByPrefixes(&flag.Prefix)
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
	return strings.Join(values, s)
}
