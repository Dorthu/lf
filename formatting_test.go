package main

import (
	"bytes"
	"io"
	"os"
	"testing"
	"text/template"
)

func TestReadFormat(t *testing.T) {
	var buf bytes.Buffer
	testData := map[string]string{
		"msg":     "it worked",
		"test":    "test",
		"example": "foobar",
	}

	// simple single-argument format
	if err := readFormat(".msg").Execute(&buf, testData); err != nil {
		t.Errorf("Error executing simple template: %v", err)
	}

	if buf.String() != "it worked" {
		t.Errorf("Simple test got %s", buf.String())
	}

	buf.Reset()

	// two-argument format
	if err := readFormat(".msg .example").Execute(&buf, testData); err != nil {
		t.Errorf("Error executing double template: %v", err)
	}

	if buf.String() != "it worked foobar" {
		t.Errorf("Double template got %s", buf.String())
	}

	buf.Reset()

	// leading and trailing characters
	if err := readFormat("[.msg] (.test) <.example>").Execute(&buf, testData); err != nil {
		t.Errorf("Error executing template: %v", err)
	}

	if buf.String() != "[it worked] (test) <foobar>" {
		t.Errorf("Leading and trialing characters got %s", buf.String())
	}

	buf.Reset()

	// missing keys
	if err := readFormat("(.missing) .msg").Execute(&buf, testData); err != nil {
		t.Errorf("Error executing missing keys template: %v", err)
	}

	if buf.String() != "() it worked" {
		t.Errorf("Missing keys test got %s", buf.String())
	}

	buf.Reset()

	// chained dots
	if err := readFormat(".msg.example").Execute(&buf, testData); err != nil {
		t.Errorf("Error executing trailing dots template: %v", err)
	}

	if buf.String() != "it worked.example" {
		t.Errorf("Trailing dots test got %s", buf.String())
	}

	buf.Reset()
}

// captureStdout returns the stdout of the evalued function
func captureStdout(f func() error) (string, error) {
	orig := os.Stdout
	out, in, _ := os.Pipe()
	os.Stdout = in

	err := f()
	if err != nil {
		return "", err
	}

	os.Stdout = orig
	in.Close()
	res, err := io.ReadAll(out)

	if err != nil {
		return "", err
	}

	return string(res), nil
}

func TestFormatLine(t *testing.T) {
	testData := map[string]string{
		"msg": "it worked",
		"key": "value",
	}
	tem, err := template.New("test").Parse("{{.msg}}")

	if err != nil {
		t.Errorf("Failed to parse test template: %v", err)
	}

	if res, err := captureStdout(
		func() error {
			formatLine(testData, tem)
			return nil
		},
	); err != nil {
		t.Errorf("Error running formatLine: %v", err)
	} else if res != "it worked\n" {
		t.Errorf("formatLine got unexpected result: '%s'", res)
	}
}

func TestDump(t *testing.T) {
	testData := map[string]string{
		"msg": "it worked",
		"key": "value",
	}

	if res, err := captureStdout(
		func() error {
			dump(testData)
			return nil
		},
	); err != nil {
		t.Errorf("Error running formatLine: %v", err)
	} else if res != `msg="it worked" key=value`+"\n" {
		t.Errorf("dump got unexpected result: '%s'", res)
	}
}
