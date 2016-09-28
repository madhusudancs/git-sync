/*
Copyright 2015 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

const (
	testKey = "KEY"
)

func TestEnvBool(t *testing.T) {
	cases := []struct {
		value string
		def   bool
		exp   bool
	}{
		{"true", true, true},
		{"true", false, true},
		{"", true, true},
		{"", false, false},
		{"false", true, false},
		{"false", false, false},
		{"", true, true},
		{"", false, false},
		{"no true", true, true},
		{"no false", true, true},
	}

	for _, testCase := range cases {
		os.Setenv(testKey, testCase.value)
		val := envBool(testKey, testCase.def)
		if val != testCase.exp {
			t.Fatalf("expected %v but %v returned", testCase.exp, val)
		}
	}
}

func TestEnvString(t *testing.T) {
	cases := []struct {
		value string
		def   string
		exp   string
	}{
		{"true", "true", "true"},
		{"true", "false", "true"},
		{"", "true", "true"},
		{"", "false", "false"},
		{"false", "true", "false"},
		{"false", "false", "false"},
		{"", "true", "true"},
		{"", "false", "false"},
	}

	for _, testCase := range cases {
		os.Setenv(testKey, testCase.value)
		val := envString(testKey, testCase.def)
		if val != testCase.exp {
			t.Fatalf("expected %v but %v returned", testCase.exp, val)
		}
	}
}

func TestEnvInt(t *testing.T) {
	cases := []struct {
		value string
		def   int
		exp   int
	}{
		{"0", 1, 0},
		{"", 0, 0},
		{"-1", 0, -1},
		{"abcd", 0, 0},
		{"abcd", 1, 1},
	}

	for _, testCase := range cases {
		os.Setenv(testKey, testCase.value)
		val := envInt(testKey, testCase.def)
		if val != testCase.exp {
			t.Fatalf("expected %v but %v returned", testCase.exp, val)
		}
	}
}

func TestNotifySync(t *testing.T) {
	local := "a9d7e156339f7dae59a32044db4d4023b4caa978"
	remote := "49112fc1876536980876d428a33a73d7a5af4156"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "HTTP Method %q not supported", http.StatusBadRequest)
			return
		}

		dec := json.NewDecoder(r.Body)
		msg := &notification{}
		err := dec.Decode(msg)
		if err != nil {
			http.Error(w, "Couldn't decode the request", http.StatusInternalServerError)
			return
		}

		if msg.LocalRev != local {
			http.Error(w, fmt.Sprintf("Unexpected local rev - got: %q, want: %q", msg.LocalRev, local), http.StatusBadRequest)
			return
		}
		if msg.RemoteRev != remote {
			http.Error(w, fmt.Sprintf("Unexpected remote rev - got: %q, want: %q", msg.RemoteRev, remote), http.StatusBadRequest)
			return
		}

		fmt.Fprintln(w, "Success!")
	}))
	defer ts.Close()

	err := notifySync(ts.URL, local, remote)
	if err != nil {
		t.Errorf("failed to notify: %v", err)
		return
	}
}
