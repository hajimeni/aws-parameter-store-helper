package client

import (
	"testing"
)

func TestRenderTemplate(t *testing.T) {
	t.Log("Start TestRenderTamplate")
	variables := []KeyValue {
		KeyValue{"key_1", "value_1"},
		KeyValue{"path/hoge-key_2", "value\"_2"},
		KeyValue{"path/hoge-key_3", `a[b+c="$\`},
	}

	flag := LoadFlag{}
	flag.Template = "export {{ .Name }}=\"{{ .Value }}\""
	flag.ReplaceKeyValue = "_"
	flag.ReplaceKeys = "-/"
	flag.UpperCaseKey = true
	flag.QuoteShell = true
	flag.Delimiter = ";"

	result1 := renderTemplate(&variables, &flag)

	if result1 != `export KEY_1="value_1";export PATH_HOGE_KEY_2="value\"_2";export PATH_HOGE_KEY_3="a\[b+c=\"\$\\"` {
		t.Fatalf("Default rendering error: %s", result1)
	}

	flag.Template = "-"
	result2 := renderTemplate(&variables, &flag)
	if result2 != "-;-;-" {
		t.Fatalf("Template no variables: %s", result2)
	}

	flag.Template = "export {{ .Name }}=\"{{ .Value }}\""
	flag.QuoteShell = false
	result3 := renderTemplate(&variables, &flag)
	if result3 != `export KEY_1="value_1";export PATH_HOGE_KEY_2="value"_2";export PATH_HOGE_KEY_3="a[b+c="$\"` {
		t.Fatalf("No Unquote: %s", result3)
	}

	flag.Template = "-D{{ .Name }}={{ .Value }}"
	flag.Delimiter = " "
	flag.NoQuoteShell = true
	result4 := renderTemplate(&variables, &flag)
	if result4 != `-DKEY_1=value_1 -DPATH_HOGE_KEY_2=value"_2 -DPATH_HOGE_KEY_3=a[b+c="$\` {
		t.Fatalf("Java args: %s", result4)
	}

}
