package main

import (
	"os"
	"testing"
)

var defaultAttr = map[string]interface{}{
	"test2": "yolo2",
}
var expectedAttr = map[string]interface{}{
	"yolo":  "test",
	"test":  "yolo",
	"test2": "yolo2",
}

var base64EnvVar = "base64,eyJ5b2xvIjoidGVzdCIsInRlc3QiOiJ5b2xvIn0="
var jsonEnvVar = `{"yolo":"test","test":"yolo"}`

func Test_overrideWithJsonIfNeeded_Base64(t *testing.T) {
	os.Setenv("Base64Var", base64EnvVar)
	result := overrideWithJsonIfNeeded("Base64Var", defaultAttr)
	if result["yolo"] != expectedAttr["yolo"] ||
		result["yolo"] != "test" {
		t.Fatalf("Expected %v, got %v", expectedAttr, result)
	}
}

func Test_overrideWithJsonIfNeeded_Json(t *testing.T) {
	os.Setenv("JsonEnvVar", jsonEnvVar)
	result := overrideWithJsonIfNeeded("JsonEnvVar", defaultAttr)
	if result["yolo"] != expectedAttr["yolo"] {
		t.Fatalf("Expected %v, got %v", expectedAttr, result)
	}
}
