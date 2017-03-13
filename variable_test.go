package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertDig(t *testing.T, v *Variable, expected, tmp interface{}, name, expr string) {
	res, err := v.Dig(tmp, name, expr)
	assert.NoError(t, err)
	assert.Equal(t, expected, res)
}

func assertExpand(t *testing.T, v *Variable, expected, expr string) {
	res, err := v.Expand(expr)
	assert.NoError(t, err)
	assert.Equal(t, expected, res)
}

func TestVariableExpandCase1(t *testing.T) {
	bucket := "bucket1"
	downloads_dir := "/tmp/workspace/downloads"
	download_files := map[string]interface{}{
		"foo": fmt.Sprintf("gs://%v/path/to/foo", bucket),
		"bar": fmt.Sprintf("gs://%v/path/to/bar", bucket),
	}
	local_download_files := map[string]interface{}{
		"foo": fmt.Sprintf("%v/%v/path/to/foo", downloads_dir, bucket),
		"bar": fmt.Sprintf("%v/%v/path/to/bar", downloads_dir, bucket),
	}

	download_files_json, err := json.Marshal(download_files)
	assert.NoError(t, err)
	local_download_files_json, err := json.Marshal(local_download_files)
	assert.NoError(t, err)

	attrs := map[string]interface{}{
		"download_files": string(download_files_json),
		"baz":            "60",
		"qux":            "data1 data2 data3",
	}
	seed := map[string]interface{}{
		"downloads_dir":  downloads_dir,
		"uploads_dir":    "/tmp/workspace/uploads",
		"download_files": string(local_download_files_json),
		"attributes":     attrs,
		"attrs":          attrs,
		"data":           "",
	}

	v := &Variable{Data: seed, Separator: " "}
	assertDig(t, v, local_download_files, seed, "download_files", "download_files")
	assertDig(t, v, local_download_files["foo"].(string), local_download_files, "foo", "foo")
	assertDig(t, v, local_download_files["bar"].(string), local_download_files, "bar", "bar")

	assertExpand(t, v, local_download_files["foo"].(string), "%{download_files.foo}")
	assertExpand(t, v, local_download_files["bar"].(string), "%{download_files.bar}")
	assertExpand(t, v, download_files["foo"].(string), "%{attrs.download_files.foo}")
	assertExpand(t, v, download_files["bar"].(string), "%{attrs.download_files.bar}")
}

func TestVariableExpandCase2(t *testing.T) {
	bucket := "bucket1"
	remote_qux := []string{
		fmt.Sprintf("gs://%v/path/to/qux1", bucket),
		fmt.Sprintf("gs://%v/path/to/qux2", bucket),
	}
	downloads_dir := "/tmp/workspace/downloads"
	download_files := map[string]interface{}{
		"bar": fmt.Sprintf("gs://%v/path/to/bar", bucket),
		"baz": fmt.Sprintf("gs://%v/path/to/baz", bucket),
		"qux": remote_qux,
	}
	local_qux := []string{
		fmt.Sprintf("%v/%v/path/to/qux1", downloads_dir, bucket),
		fmt.Sprintf("%v/%v/path/to/qux2", downloads_dir, bucket),
	}
	local_download_files := map[string]interface{}{
		"bar": fmt.Sprintf("%v/%v/path/to/bar", downloads_dir, bucket),
		"baz": fmt.Sprintf("%v/%v/path/to/baz", downloads_dir, bucket),
		"qux": local_qux,
	}

	download_files_json, err := json.Marshal(download_files)
	assert.NoError(t, err)
	local_download_files_json, err := json.Marshal(local_download_files)
	assert.NoError(t, err)

	attrs := map[string]interface{}{
		"download_files": string(download_files_json),
		"foo":            "123",
	}
	seed := map[string]interface{}{
		"downloads_dir":  downloads_dir,
		"uploads_dir":    "/tmp/workspace/uploads",
		"download_files": string(local_download_files_json),
		"attributes":     attrs,
		"attrs":          attrs,
		"data":           "",
	}

	v := &Variable{Data: seed, Separator: " "}
	assertExpand(t, v, "123", "%{attrs.foo}")
	assertExpand(t, v, strings.Join(local_qux, " "), "%{download_files.qux}")
	assertExpand(t, v, strings.Join(remote_qux, " "), "%{attrs.download_files.qux}")
}
