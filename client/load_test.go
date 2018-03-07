package client

import (
	"testing"
)

func TestRenderTemplate(t *testing.T) {
	t.Log("Start TestRenderTamplate")
	variables := []KeyValue {
		KeyValue{"key_1", "value_1"},
		KeyValue{"path/hoge-key_2", "value\"_2"},
	}

	flag := LoadFlag{}
	flag.Template = "export {{ .Name }}=\"{{ .Value }}\""
	flag.ReplaceKeyValue = "_"
	flag.ReplaceKeys = "-/"
	flag.UpperCaseKey = true
	flag.EscapeDoublequote = "\\"
	flag.Delimiter = ";"

	result1 := renderTemplate(&variables, &flag)

	if result1 != "export KEY_1=\"value_1\";export PATH_HOGE_KEY_2=\"value\\\"_2\"" {
		t.Fatal("Default rendering error: %s", result1)
	}

	flag.Template = "-"
	result2 := renderTemplate(&variables, &flag)
	if result2 != "-;-" {
		t.Fatal("Template no variagles: %s", result2)
	}

}
