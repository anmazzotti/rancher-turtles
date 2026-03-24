/*
Copyright © 2026 SUSE LLC

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

package testenv

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	turtlesframework "github.com/rancher/turtles/test/framework"
)

// RunCollectETCDDataInput is the input for RunCollectETCDData
type RunCollectETCDDataInput struct {
	// ClusterName is the cluster for which the data is collected.
	// Beware that launching the script multiple times for the same cluster will override the data.
	ClusterName string

	// ContainerName is the container to use to run the collection script.
	ContainerName string

	// CreateRancherCertsScriptPath is the file path to the certs creation script.
	CollectETCDDataScriptPath string `env:"ETCD_DUMP_SCRIPT_PATH"`

	// CreateRancherCertsScriptPath is the file path to the certs creation script.
	ETCDDumpDirectory string `env:"ETCD_DUMP_DIRECTORY"`

	// ETCDRunCollection is the flag used to determine whether to run or skip collection.
	// This is useful to skip execution when the management cluster is not supported, ex. EKS.
	ETCDRunCollection string `env:"ETCD_RUN_COLLECTION"`

	// ArtifactsFolder is the root path for the artifacts
	ArtifactsFolder string `env:"ARTIFACTS_FOLDER"`
}

// RunCollectETCDData runs etcd data collection and size verification for a given cluster.
// Note that this function uses 'docker exec', expecting the script to already be present
// within the targeted container.
func RunCollectETCDData(ctx context.Context, input RunCollectETCDDataInput) {
	Expect(turtlesframework.Parse(&input)).To(Succeed(), "Failed to parse environment variables")

	if input.ETCDRunCollection != "true" {
		GinkgoWriter.Printf("Skipping ETCD data collection")
		return
	}

	Expect(ctx).ShouldNot(BeNil(), "ctx is required for RunCollectETCDData")
	Expect(input.ClusterName).ShouldNot(BeEmpty(), "ClusterName is required for RunCollectETCDData")
	Expect(input.ContainerName).ShouldNot(BeEmpty(), "ContainerName is required for RunCollectETCDData")
	Expect(input.CollectETCDDataScriptPath).ShouldNot(BeEmpty(), "CollectETCDDataScriptPath is required for RunCollectETCDData")
	Expect(input.ETCDDumpDirectory).ShouldNot(BeEmpty(), "ETCDDumpDirectory is required for RunCollectETCDData")
	Expect(input.ArtifactsFolder).ShouldNot(BeEmpty(), "ArtifactsFolder is required for RunCollectETCDData")

	collectETCDDataResult := turtlesframework.RunCommand(ctx, turtlesframework.RunCommandInput{
		Command: "docker",
		Args: []string{
			"exec",
			"--env", fmt.Sprintf("CLUSTER_NAME=%s", input.ClusterName),
			input.ContainerName,
			input.CollectETCDDataScriptPath,
		},
	})

	// Store any dumped info into artifacts before checking for errors.
	etcdArtifactsFolder := input.ArtifactsFolder + "/etcd"
	turtlesframework.RunCommand(ctx, turtlesframework.RunCommandInput{
		Command: "mkdir",
		Args: []string{
			"-p",
			etcdArtifactsFolder,
		},
	})
	turtlesframework.RunCommand(ctx, turtlesframework.RunCommandInput{
		Command: "mv",
		Args: []string{
			input.ETCDDumpDirectory + "/*.txt",
			etcdArtifactsFolder,
		},
	})

	Expect(collectETCDDataResult.Error).ShouldNot(HaveOccurred(), "Failed collecting ETCD data")
	Expect(collectETCDDataResult.ExitCode).To(Equal(0), "Collecting ETCD data returned non-zero exit code")
	if collectETCDDataResult.ExitCode != 0 {
		GinkgoWriter.Printf("Error: %s\n", string(collectETCDDataResult.Stderr))
	}

}
