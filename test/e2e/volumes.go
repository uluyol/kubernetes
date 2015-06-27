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

/*
 * This test checks that various VolumeSources are working. For each volume
 * type it creates a server pod, exporting simple 'index.html' file.
 * Then it uses appropriate VolumeSource to import this file into a client pod
 * and checks that the pod can see the file. It does so by importing the file
 * into web server root and loadind the index.html from it.
 *
 * These tests work only when privileged containers are allowed, exporting
 * various filesystems (NFS, GlusterFS, ...) usually needs some mounting or
 * other privileged magic in the server pod.
 *
 * Note that the server containers are for testing purposes only and should not
 * be used in production.
 */

package e2e

import (
	"fmt"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/client"
	"github.com/GoogleCloudPlatform/kubernetes/pkg/runtime"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// Configuration of one tests. The test consist of:
// - server pod - runs serverImage, exports ports[]
// - client pod - does not need any special configuration
type VolumeTestConfig struct {
	namespace string
	// Prefix of all pods. Typically the test name.
	prefix string
	// Name of container image for the server pod.
	serverImage string
	// Ports to export from the server pod. TCP only.
	serverPorts []int
}

// Starts a container specified by config.serverImage and exports all
// config.serverPorts from it. The returned pod should be used to get the server
// IP address and create appropriate VolumeSource.
func startVolumeServer(client *client.Client, config VolumeTestConfig) *api.Pod {
	podClient := client.Pods(config.namespace)

	portCount := len(config.serverPorts)
	serverPodPorts := make([]api.ContainerPort, portCount)

	for i := 0; i < portCount; i++ {
		portName := fmt.Sprintf("%s-%d", config.prefix, i)

		serverPodPorts[i] = api.ContainerPort{
			Name:          portName,
			ContainerPort: config.serverPorts[i],
			Protocol:      api.ProtocolTCP,
		}
	}

	By(fmt.Sprint("creating ", config.prefix, " server pod"))
	privileged := new(bool)
	*privileged = true
	serverPod := &api.Pod{
		TypeMeta: runtime.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name: config.prefix + "-server",
			Labels: map[string]string{
				"role": config.prefix + "-server",
			},
		},

		Spec: api.PodSpec{
			Containers: []api.Container{
				{
					Name:  config.prefix + "-server",
					Image: config.serverImage,
					SecurityContext: &api.SecurityContext{
						Privileged: privileged,
					},
					Ports: serverPodPorts,
				},
			},
		},
	}
	_, err := podClient.Create(serverPod)
	expectNoError(err, "Failed to create %s pod: %v", serverPod.Name, err)

	expectNoError(waitForPodRunningInNamespace(client, serverPod.Name, config.namespace))

	By("locating the server pod")
	pod, err := podClient.Get(serverPod.Name)
	expectNoError(err, "Cannot locate the server pod %v: %v", serverPod.Name, err)

	By("sleeping a bit to give the server time to start")
	time.Sleep(20 * time.Second)
	return pod
}

// Clean both server and client pods.
func volumeTestCleanup(client *client.Client, config VolumeTestConfig) {
	By(fmt.Sprint("cleaning the environment after ", config.prefix))

	defer GinkgoRecover()

	podClient := client.Pods(config.namespace)

	// ignore all errors, the pods may not be even created
	podClient.Delete(config.prefix+"-client", nil)
	podClient.Delete(config.prefix+"-server", nil)
}

// Start a client pod using given VolumeSource (exported by startVolumeServer())
// and check that the pod sees the data from the server pod.
func testVolumeClient(client *client.Client, config VolumeTestConfig, volume api.VolumeSource, expectedContent string) {
	By(fmt.Sprint("starting ", config.prefix, " client"))
	podClient := client.Pods(config.namespace)

	clientPod := &api.Pod{
		TypeMeta: runtime.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: api.ObjectMeta{
			Name: config.prefix + "-client",
			Labels: map[string]string{
				"role": config.prefix + "-client",
			},
		},
		Spec: api.PodSpec{
			Containers: []api.Container{
				{
					Name:  config.prefix + "-client",
					Image: "gcr.io/google_containers/nginx:1.7.9",
					Ports: []api.ContainerPort{
						{
							Name:          "web",
							ContainerPort: 80,
							Protocol:      api.ProtocolTCP,
						},
					},
					VolumeMounts: []api.VolumeMount{
						{
							Name:      config.prefix + "-volume",
							MountPath: "/usr/share/nginx/html",
						},
					},
				},
			},
			Volumes: []api.Volume{
				{
					Name:         config.prefix + "-volume",
					VolumeSource: volume,
				},
			},
		},
	}
	if _, err := podClient.Create(clientPod); err != nil {
		Failf("Failed to create %s pod: %v", clientPod.Name, err)
	}
	expectNoError(waitForPodRunningInNamespace(client, clientPod.Name, config.namespace))

	By("reading a web page from the client")
	body, err := client.Get().
		Namespace(config.namespace).
		Prefix("proxy").
		Resource("pods").
		Name(clientPod.Name).
		DoRaw()
	expectNoError(err, "Cannot read web page: %v", err)
	Logf("body: %v", string(body))

	By("checking the page content")
	Expect(body).To(ContainSubstring(expectedContent))
}

var _ = Describe("Volumes", func() {
	clean := true // If 'false', the test won't clear its namespace (and pods and services) upon completion. Useful for debugging.

	// filled in BeforeEach
	var c *client.Client
	var namespace *api.Namespace

	BeforeEach(func() {
		var err error
		c, err = loadClient()
		Expect(err).NotTo(HaveOccurred())
		By("Building a namespace api object")
		namespace, err = createTestingNS("volume", c)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if clean {
			if err := c.Namespaces().Delete(namespace.Name); err != nil {
				Failf("Couldn't delete ns %s", err)
			}
		}
	})

	////////////////////////////////////////////////////////////////////////
	// NFS
	////////////////////////////////////////////////////////////////////////

	// Marked with [Skipped] to skip the test by default (see driver.go),
	// the test needs privileged containers, which are disabled by default.
	// Run the test with "go run hack/e2e.go ... --ginkgo.focus=Volume"
	Describe("[Skipped] NFS", func() {
		It("should be mountable", func() {
			config := VolumeTestConfig{
				namespace:   namespace.Name,
				prefix:      "nfs",
				serverImage: "gcr.io/google_containers/volume-nfs",
				serverPorts: []int{2049},
			}

			defer func() {
				if clean {
					volumeTestCleanup(c, config)
				}
			}()
			pod := startVolumeServer(c, config)
			serverIP := pod.Status.PodIP
			Logf("NFS server IP address: %v", serverIP)

			volume := api.VolumeSource{
				NFS: &api.NFSVolumeSource{
					Server:   serverIP,
					Path:     "/",
					ReadOnly: true,
				},
			}
			// Must match content of contrib/for-tests/volumes-tester/nfs/index.html
			testVolumeClient(c, config, volume, "Hello from NFS!")
		})
	})

	////////////////////////////////////////////////////////////////////////
	// Gluster
	////////////////////////////////////////////////////////////////////////

	// Marked with [Skipped] to skip the test by default (see driver.go),
	// the test needs privileged containers, which are disabled by default.
	// Run the test with "go run hack/e2e.go ... --ginkgo.focus=Volume"
	Describe("[Skipped] GlusterFS", func() {
		It("should be mountable", func() {
			config := VolumeTestConfig{
				namespace:   namespace.Name,
				prefix:      "gluster",
				serverImage: "gcr.io/google_containers/volume-gluster",
				serverPorts: []int{24007, 24008, 49152},
			}

			defer func() {
				if clean {
					volumeTestCleanup(c, config)
				}
			}()
			pod := startVolumeServer(c, config)
			serverIP := pod.Status.PodIP
			Logf("Gluster server IP address: %v", serverIP)

			// create Endpoints for the server
			endpoints := api.Endpoints{
				TypeMeta: runtime.TypeMeta{
					Kind:       "Endpoints",
					APIVersion: "v1",
				},
				ObjectMeta: api.ObjectMeta{
					Name: config.prefix + "-server",
				},
				Subsets: []api.EndpointSubset{
					{
						Addresses: []api.EndpointAddress{
							{
								IP: serverIP,
							},
						},
						Ports: []api.EndpointPort{
							{
								Name:     "gluster",
								Port:     24007,
								Protocol: api.ProtocolTCP,
							},
						},
					},
				},
			}

			endClient := c.Endpoints(config.namespace)

			defer func() {
				if clean {
					endClient.Delete(config.prefix + "-server")
				}
			}()

			if _, err := endClient.Create(&endpoints); err != nil {
				Failf("Failed to create endpoints for Gluster server: %v", err)
			}

			volume := api.VolumeSource{
				Glusterfs: &api.GlusterfsVolumeSource{
					EndpointsName: config.prefix + "-server",
					// 'test_vol' comes from contrib/for-tests/volumes-tester/gluster/run_gluster.sh
					Path:     "test_vol",
					ReadOnly: true,
				},
			}
			// Must match content of contrib/for-tests/volumes-tester/gluster/index.html
			testVolumeClient(c, config, volume, "Hello from GlusterFS!")
		})
	})
})
