package e2e

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mittwald/go-helm-client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	clientset "k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultKubeConfigPath = ".kube/config"
)

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Suite")
}

// framework is container for components can be reused for each test
type framework struct {
	clientConfig *rest.Config
	clientSet    clientset.Interface
	namespace    string
	helmClient   helmclient.Client
	helmChart    *helmclient.ChartSpec
}

func abs(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		panic("check path")
	}

	return absPath
}

func newFrameworkWithDefaultNamespace() (*framework, error) {
	var defaultNS = "kube-system"

	homePath, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	kubeconfig := homePath + "/" + defaultKubeConfigPath
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	helmChartClient, err := helmclient.NewClientFromRestConf(&helmclient.RestConfClientOptions{
		Options: &helmclient.Options{
			Namespace: defaultNS,
		},
		RestConfig: config,
	})
	if err != nil {
		return nil, err
	}

	return &framework{
		clientConfig: config,
		clientSet:    clientSet,
		helmClient:   helmChartClient,
		namespace:    defaultNS,
		helmChart:    nil,
	}, nil

}

var frk *framework

// TODO(@bg): we may need to set up kubernetes cluster in e2e-test to run test for supported versions
var _ = BeforeSuite(func() {
	newFrk, err := newFrameworkWithDefaultNamespace()
	Expect(err).To(BeNil())
	frk = newFrk

	// list namespaces to ensure api-server accessibility
	list, err := frk.clientSet.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	Expect(err).To(BeNil())
	Expect(len(list.Items)).Should(BeNumerically(">=", 1))
})

// Note(@bg): assumption is that this test will be run on the two socket workstation with two NPUs per each socket.
// TODO: enhance test to parse resource name from node object
var _ = Describe("end-to-end test", func() {
	Context("test legacy strategy", func() {
		It("deploy device-plugin helm chart for legacy strategy", deployHelmChart("legacy"))

		It("verify node", verifyNode("alpha.furiosa.ai/npu"))

		It("request NPUs", deployVerificationPod("alpha.furiosa.ai/npu"))

		It("clean up verification pod", cleanUpVerificationPod())

		It("verify pod environment", verifyInferenceEnv("alpha.furiosa.ai/npu"))

		It("clean up verification pod", cleanUpInferencePod())

		It("delete helm chart", deleteHelmChart())
	})

	Context("test generic strategy ", func() {
		It("deploy device-plugin helm chart for legacy strategy", deployHelmChart("generic"))

		It("verify node", verifyNode("furiosa.ai/warboy"))

		It("request NPUs", deployVerificationPod("furiosa.ai/warboy"))

		It("clean up verification pod", cleanUpVerificationPod())

		It("verify pod environment", verifyInferenceEnv("furiosa.ai/warboy"))

		It("clean up verification pod", cleanUpInferencePod())

		It("delete helm chart", deleteHelmChart())
	})

	//TODO: add more tests
})

func genVerificationPodManifest(npuNum string, resourceName string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "verification-pod",
			Namespace: frk.namespace,
		},
		Spec: v1.PodSpec{
			RestartPolicy: v1.RestartPolicyNever,
			Containers: []v1.Container{
				{
					Name:            "verification-pod",
					Image:           "ghcr.io/furiosa-ai/furiosa-device-plugin/e2e/verification:latest",
					ImagePullPolicy: v1.PullAlways,
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceName(resourceName): resource.MustParse(npuNum),
						},
					},
				},
			},
			Tolerations: []v1.Toleration{
				{
					Key:      "npu",
					Operator: v1.TolerationOpExists,
				},
			},
		},
	}
}

func genInferencePodManifest(resourceName string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "inference-pod",
			Namespace: frk.namespace,
		},
		Spec: v1.PodSpec{
			RestartPolicy: v1.RestartPolicyNever,
			Containers: []v1.Container{
				{
					Name:            "inference-pod",
					Image:           "ghcr.io/furiosa-ai/furiosa-device-plugin/e2e/inference:latest",
					ImagePullPolicy: v1.PullAlways,
					Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							v1.ResourceName(resourceName): resource.MustParse("1"),
						},
					},
				},
			},
			Tolerations: []v1.Toleration{
				{
					Key:      "npu",
					Operator: v1.TolerationOpExists,
				},
			},
		},
	}
}

func composeValues(strategy string) string {
	template := `namespace: kube-system
daemonSet:
  priorityClassName: system-node-critical
  # Use OnDelete for change the plugin strategy
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
  tolerations:
    - key: npu
      operator: Exists
  image:
    repository: ghcr.io/furiosa-ai/furiosa-device-plugin
    tag: latest
    pullPolicy: Always
  resources:
    cpu: 100m
    memory: 64Mi

globalConfig:
  resourceStrategyMap:
    warboy: %s
    rngd: %s
  debugMode: false
`
	return fmt.Sprintf(template, strategy, strategy)
}

func strRand() string {
	return fmt.Sprintf("%d", rand.Int())
}

func deployHelmChart(strategy string) func() {
	return func() {
		helmChartSpec := &helmclient.ChartSpec{
			ReleaseName:     "e2e-test" + strRand(),
			ChartName:       abs("../deployments/helm"), //path to helm chart
			Namespace:       frk.namespace,
			CreateNamespace: false,
			Wait:            false,
			Timeout:         5 * time.Minute,
			CleanupOnFail:   false,
			ValuesYaml:      composeValues(strategy),
		}
		frk.helmChart = helmChartSpec

		_, err := frk.helmClient.InstallChart(context.TODO(), frk.helmChart, nil)
		Expect(err).To(BeNil())
	}
}

func verifyNode(resUniqueKeys ...string) func() {
	return func() {
		//TODO: check each node with specific taint has designated privileged daemon
		nodeList, err := frk.clientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		Expect(err).To(BeNil())
		Expect(len(nodeList.Items)).Should(BeNumerically(">=", 1))

		nodeName := nodeList.Items[0].Name
		podList, err := frk.clientSet.CoreV1().Pods(frk.namespace).List(context.TODO(), metav1.ListOptions{
			FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName),
			LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s", "furiosa-device-plugin"),
		})
		Expect(err).To(BeNil())
		Expect(len(podList.Items)).Should(BeNumerically("==", 1))

		// polling until pod.status.phase of daemonset became Running with timeout 15 sec
		podName := podList.Items[0].Name
		Eventually(func() v1.PodPhase {
			pod, err := frk.clientSet.CoreV1().Pods(frk.namespace).Get(context.TODO(), podName, metav1.GetOptions{})
			if err != nil {
				return v1.PodUnknown
			}
			return pod.Status.Phase
		}).WithPolling(1 * time.Second).WithTimeout(15 * time.Second).Should(Equal(v1.PodRunning))

		// polling the same node for resource name and quantity verification
		Eventually(func() int {
			node, err := frk.clientSet.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
			if err != nil {
				return 0
			}

			var foundKeys []string
			for _, resUniqueKey := range resUniqueKeys {
				capacity := node.Status.Capacity[v1.ResourceName(resUniqueKey)]
				allocatable := node.Status.Allocatable[v1.ResourceName(resUniqueKey)]

				decimalCapacity, _ := capacity.AsInt64()
				decimalAllocatable, _ := allocatable.AsInt64()

				if decimalCapacity > 0 && decimalCapacity == decimalAllocatable {
					foundKeys = append(foundKeys, resUniqueKey)
				}
			}

			return len(foundKeys)
		}).WithPolling(1 * time.Second).WithTimeout(15 * time.Second).Should(BeNumerically(">=", 1))
	}
}

func deployVerificationPod(resourceName string) func() {
	return func() {
		// deploy verification pod
		_, err := frk.clientSet.CoreV1().Pods(frk.namespace).Create(context.TODO(), genVerificationPodManifest("1", resourceName), metav1.CreateOptions{})
		Expect(err).To(BeNil())

		// polling until pod.status.phase became succeeded with timeout 30 sec
		Eventually(func() v1.PodPhase {
			pod, err := frk.clientSet.CoreV1().Pods(frk.namespace).Get(context.TODO(), "verification-pod", metav1.GetOptions{})
			if err != nil {
				return v1.PodUnknown
			}
			return pod.Status.Phase
		}).WithPolling(1 * time.Second).WithTimeout(30 * time.Second).Should(Equal(v1.PodSucceeded))

		// parse allocated npu list through CoreV1().Pods().GetLogs() api
		request := frk.clientSet.CoreV1().Pods(frk.namespace).GetLogs("verification-pod", &v1.PodLogOptions{})
		logs, err := request.Stream(context.TODO())
		Expect(err).To(BeNil())
		defer logs.Close()

		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, logs)
		Expect(err).To(BeNil())

		devices, err := UnmarshalDevices(buf.Bytes())
		Expect(err).To(BeNil())
		Expect(len(devices.Devices)).Should(BeNumerically(">=", 1))
	}
}

func cleanUpVerificationPod() func() {
	return func() {
		// delete verification pod
		err := frk.clientSet.CoreV1().Pods(frk.namespace).Delete(context.TODO(), "verification-pod", metav1.DeleteOptions{})
		Expect(err).To(BeNil())
	}
}

func verifyInferenceEnv(resourceName string) func() {
	return func() {
		// deploy inference pod
		_, err := frk.clientSet.CoreV1().Pods(frk.namespace).Create(context.TODO(), genInferencePodManifest(resourceName), metav1.CreateOptions{})
		Expect(err).To(BeNil())

		// polling until pod.status.phase became succeeded with timeout up to 5 min since image size is bigger than 5GB
		Eventually(func() v1.PodPhase {
			pod, err := frk.clientSet.CoreV1().Pods(frk.namespace).Get(context.TODO(), "inference-pod", metav1.GetOptions{})
			if err != nil {
				return v1.PodUnknown
			}
			return pod.Status.Phase
		}).WithPolling(1 * time.Second).WithTimeout(300 * time.Second).Should(Equal(v1.PodSucceeded))
	}
}

func cleanUpInferencePod() func() {
	return func() {
		// delete inference pod
		err := frk.clientSet.CoreV1().Pods(frk.namespace).Delete(context.TODO(), "inference-pod", metav1.DeleteOptions{})
		Expect(err).To(BeNil())
	}
}

func deleteHelmChart() func() {
	return func() {
		err := frk.helmClient.UninstallRelease(frk.helmChart)
		Expect(err).To(BeNil())
	}
}
