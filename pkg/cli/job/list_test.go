/*
Copyright 2019 The Volcano Authors.

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

package job

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"volcano.sh/volcano/pkg/cli/util"

	"github.com/spf13/cobra"

	"volcano.sh/apis/pkg/apis/batch/v1alpha1"
)

func TestListJob(t *testing.T) {
	response := v1alpha1.JobList{}
	response.Items = append(response.Items, v1alpha1.Job{})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		val, err := json.Marshal(response)
		if err == nil {
			w.Write(val)
		}

	})

	server := httptest.NewServer(handler)
	defer server.Close()

	testCases := []struct {
		Name         string
		ExpectValue  error
		AllNamespace bool
		Selector     string
	}{
		{
			Name:        "ListJob",
			ExpectValue: nil,
		},
		{
			Name:         "ListAllNamespaceJob",
			ExpectValue:  nil,
			AllNamespace: true,
		},
	}

	for i, testcase := range testCases {
		listJobFlags = &listFlags{
			CommonFlags: util.CommonFlags{
				Master: server.URL,
			},
			Namespace:    "test",
			allNamespace: testcase.AllNamespace,
			selector:     testcase.Selector,
		}

		err := ListJobs(context.TODO())
		if err != nil {
			t.Errorf("case %d (%s): expected: %v, got %v ", i, testcase.Name, testcase.ExpectValue, err)
		}
	}

}

func TestInitListFlags(t *testing.T) {
	var cmd cobra.Command
	InitListFlags(&cmd)

	if cmd.Flag("namespace") == nil {
		t.Errorf("Could not find the flag namespace")
	}
	if cmd.Flag("scheduler") == nil {
		t.Errorf("Could not find the flag scheduler")
	}
	if cmd.Flag("all-namespaces") == nil {
		t.Errorf("Could not find the flag all-namespaces")
	}
	if cmd.Flag("selector") == nil {
		t.Errorf("Could not find the flag selector")
	}

}
