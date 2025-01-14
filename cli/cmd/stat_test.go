package cmd

import (
	"testing"

	"github.com/linkerd/linkerd2/controller/api/public"
	"github.com/linkerd/linkerd2/pkg/k8s"
)

type paramsExp struct {
	counts  *public.PodCounts
	options *statOptions
	resNs   []string
	file    string
}

func TestStat(t *testing.T) {
	options := newStatOptions()
	t.Run("Returns namespace stats", func(t *testing.T) {
		testStatCall(paramsExp{
			counts: &public.PodCounts{
				MeshedPods:  1,
				RunningPods: 2,
				FailedPods:  0,
			},
			options: options,
			resNs:   []string{"emojivoto1"},
			file:    "stat_one_output.golden",
		}, k8s.Namespace, t)
	})

	t.Run("Returns pod stats", func(t *testing.T) {
		testStatCall(paramsExp{
			counts: &public.PodCounts{
				Status:      "Running",
				MeshedPods:  1,
				RunningPods: 1,
				FailedPods:  0,
			},
			options: options,
			resNs:   []string{"emojivoto1"},
			file:    "stat_one_pod_output.golden",
		}, k8s.Pod, t)
	})

	options.outputFormat = jsonOutput
	t.Run("Returns namespace stats (json)", func(t *testing.T) {
		testStatCall(paramsExp{
			counts: &public.PodCounts{
				MeshedPods:  1,
				RunningPods: 2,
				FailedPods:  0,
			},
			options: options,
			resNs:   []string{"emojivoto1"},
			file:    "stat_one_output_json.golden",
		}, k8s.Namespace, t)
	})

	options = newStatOptions()
	options.allNamespaces = true
	t.Run("Returns all namespace stats", func(t *testing.T) {
		testStatCall(paramsExp{
			counts: &public.PodCounts{
				MeshedPods:  1,
				RunningPods: 2,
				FailedPods:  0,
			},
			options: options,
			resNs:   []string{"emojivoto1", "emojivoto2"},
			file:    "stat_all_output.golden",
		}, k8s.Namespace, t)
	})

	options.outputFormat = jsonOutput
	t.Run("Returns all namespace stats (json)", func(t *testing.T) {
		testStatCall(paramsExp{
			counts: &public.PodCounts{
				MeshedPods:  1,
				RunningPods: 2,
				FailedPods:  0,
			},
			options: options,
			resNs:   []string{"emojivoto1", "emojivoto2"},
			file:    "stat_all_output_json.golden",
		}, k8s.Namespace, t)
	})

	options = newStatOptions()
	options.outputFormat = "wide"
	t.Run("Returns TCP stats", func(t *testing.T) {
		testStatCall(paramsExp{
			counts: &public.PodCounts{
				MeshedPods:  1,
				RunningPods: 2,
				FailedPods:  0,
			},
			options: options,
			resNs:   []string{"emojivoto1"},
			file:    "stat_one_tcp_output.golden",
		}, k8s.Namespace, t)
	})

	t.Run("Returns an error for named resource queries with the --all-namespaces flag", func(t *testing.T) {
		options := newStatOptions()
		options.allNamespaces = true
		args := []string{"po", "web"}
		expectedError := "stats for a resource cannot be retrieved by name across all namespaces"

		_, err := buildStatSummaryRequests(args, options)
		if err == nil || err.Error() != expectedError {
			t.Fatalf("Expected error [%s] instead got [%s]", expectedError, err)
		}
	})

	t.Run("Rejects commands with both --to and --from flags", func(t *testing.T) {
		options := newStatOptions()
		options.toResource = "deploy/foo"
		options.fromResource = "deploy/bar"
		args := []string{"ns", "test"}
		expectedError := "--to and --from flags are mutually exclusive"

		_, err := buildStatSummaryRequests(args, options)
		if err == nil || err.Error() != expectedError {
			t.Fatalf("Expected error [%s] instead got [%s]", expectedError, err)
		}
	})

	t.Run("Rejects commands with both --to-namespace and --from-namespace flags", func(t *testing.T) {
		options := newStatOptions()
		options.toNamespace = "foo"
		options.fromNamespace = "bar"
		args := []string{"po"}
		expectedError := "--to-namespace and --from-namespace flags are mutually exclusive"

		_, err := buildStatSummaryRequests(args, options)
		if err == nil || err.Error() != expectedError {
			t.Fatalf("Expected error [%s] instead got [%s]", expectedError, err)
		}
	})

	t.Run("Rejects --to-namespace flag when the target is a namespace", func(t *testing.T) {
		options := newStatOptions()
		options.toNamespace = "bar"
		args := []string{"ns", "foo"}
		expectedError := "--to-namespace flag is incompatible with namespace resource type"

		_, err := buildStatSummaryRequests(args, options)
		if err == nil || err.Error() != expectedError {
			t.Fatalf("Expected error [%s] instead got [%s]", expectedError, err)
		}
	})

	t.Run("Rejects --from-namespace flag when the target is a namespace", func(t *testing.T) {
		options := newStatOptions()
		options.fromNamespace = "foo"
		args := []string{"ns/bar"}
		expectedError := "--from-namespace flag is incompatible with namespace resource type"

		_, err := buildStatSummaryRequests(args, options)
		if err == nil || err.Error() != expectedError {
			t.Fatalf("Expected error [%s] instead got [%s]", expectedError, err)
		}
	})
}

func testStatCall(exp paramsExp, resourceType string, t *testing.T) {
	mockClient := &public.MockAPIClient{}
	response := public.GenStatSummaryResponse("emoji", resourceType, exp.resNs, exp.counts, true, true)

	mockClient.StatSummaryResponseToReturn = &response

	args := []string{"ns"}
	reqs, err := buildStatSummaryRequests(args, exp.options)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	resp, err := requestStatsFromAPI(mockClient, reqs[0])
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	rows := respToRows(resp)
	output := renderStatStats(rows, exp.options)

	diffTestdata(t, exp.file, output)
}
