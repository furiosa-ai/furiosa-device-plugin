package e2e

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/furiosa-ai/libfuriosa-kubernetes/pkg/e2e"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

var (
	e2eTestImageRegistry string
	e2eTestImageName     string
	e2eTestImageTag      string
)

func init() {
	e2eTestImageRegistry = os.Getenv("E2E_TEST_IMAGE_REGISTRY")
	if e2eTestImageRegistry == "" {
		e2eTestImageRegistry = "registry.corp.furiosa.ai/furiosa"
	}

	e2eTestImageName = os.Getenv("E2E_TEST_IMAGE_NAME")
	if e2eTestImageName == "" {
		e2eTestImageName = "furiosa-feature-discovery"
	}

	e2eTestImageTag = os.Getenv("E2E_TEST_IMAGE_TAG")
	if e2eTestImageTag == "" {
		e2eTestImageTag = "latest"
	}
}

func TestE2E(t *testing.T) {
	e2e.GenericRunTestSuiteFunc(t, "device-plugin e2e test")
}

var _ = BeforeSuite(func() {
	e2e.GenericBeforeSuiteFunc()
})

func abs(path string) string {
	absPath, err := filepath.Abs(path)
	if err != nil {
		panic("check path")
	}

	return absPath
}

// Note(@bg): assumption is that this test will be run on the two socket workstation with two NPUs per each socket.
var _ = Describe("test legacy strategy", Ordered, func() {
	// deploy device-plugin helm chart for legacy strategy
	BeforeAll(e2e.DeployHelmChart("legacy-strategy", abs("../deployments/helm"), composeValues(e2eTestImageRegistry, e2eTestImageName, e2eTestImageTag, "legacy")))

	It("verify node", verifyNode("alpha.furiosa.ai/npu"))

	It("request NPUs", deployVerificationPodAndVerifyEnv("alpha.furiosa.ai/npu"))

	// FIXME(@bg): run inference pod once image is ready
	/*It("verify pod environment", verifyInferenceEnv("alpha.furiosa.ai/npu"))

	It("clean up verification pod", cleanUpInferencePod())*/

	AfterAll(cleanUpVerificationPodIfExist())
	AfterAll(e2e.DeleteHelmChart())
})

var _ = Describe("test generic strategy", Ordered, func() {
	// deploy device-plugin helm chart for generic strategy

	arch := os.Getenv("FURIOSA_ARCH")
	if arch == "" {
		Fail("FURIOSA_ARCH env var is not set")
		return
	}

	BeforeAll(e2e.DeployHelmChart("legacy-strategy", abs("../deployments/helm"), composeValues(e2eTestImageRegistry, e2eTestImageName, e2eTestImageTag, "generic")))

	It("verify node", verifyNode(fmt.Sprintf("furiosa.ai/%s", arch)))

	It("request NPUs", deployVerificationPodAndVerifyEnv(fmt.Sprintf("furiosa.ai/%s", arch)))

	// FIXME(@bg): run inference pod once image is ready
	/*It("verify pod environment", verifyInferenceEnv(fmt.Sprintf("furiosa.ai/%s", arch)))

	It("clean up verification pod", cleanUpInferencePod())*/

	AfterAll(cleanUpVerificationPodIfExist())
	AfterAll(e2e.DeleteHelmChart())
})

func genVerificationPodManifest(npuNum string, resourceName string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "verification-pod",
			Namespace: e2e.BackgroundContext().Namespace,
		},
		Spec: v1.PodSpec{
			RestartPolicy: v1.RestartPolicyNever,
			Containers: []v1.Container{
				{
					Name:            "verification-pod",
					Image:           "registry.corp.furiosa.ai/furiosa/furiosa-device-plugin/e2e/verification:latest",
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

/*func genInferencePodManifest(resourceName string) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "inference-pod",
			Namespace: e2e.BackgroundContext().Namespace,
		},
		Spec: v1.PodSpec{
			RestartPolicy: v1.RestartPolicyNever,
			Containers: []v1.Container{
				{
					Name:            "inference-pod",
					Image:           "registry.corp.furiosa.ai/furiosa/furiosa-device-plugin/e2e/inference:latest",
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
}*/

func composeValues(e2eTestImageRegistry, e2eTestImageName, e2eTestImageTag, strategy string) string {
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
    repository: %s/%s
    tag: %s
    pullPolicy: Always
  resources:
    cpu: 100m
    memory: 64Mi

config:
  resourceStrategy: %s
  debugMode: false
`
	return fmt.Sprintf(template, e2eTestImageRegistry, e2eTestImageName, e2eTestImageTag, strategy)
}

func verifyNode(resUniqueKeys ...string) func() {
	return func() {
		var nodeList *v1.NodeList
		var nodeName string
		Eventually(func() int {
			var err error
			nodeList, err = e2e.BackgroundContext().ClientSet.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				return 0
			}
			if len(nodeList.Items) > 0 {
				nodeName = nodeList.Items[0].Name
			}
			return len(nodeList.Items)
		}).WithPolling(1 * time.Second).WithTimeout(300 * time.Second).Should(BeNumerically(">=", 1))

		var podList *v1.PodList
		Eventually(func() int {
			var err error
			podList, err = e2e.BackgroundContext().ClientSet.CoreV1().Pods(e2e.BackgroundContext().Namespace).List(context.TODO(), metav1.ListOptions{
				FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName),
				LabelSelector: fmt.Sprintf("app.kubernetes.io/name=%s", "furiosa-device-plugin"),
			})
			if err != nil {
				return 0
			}
			return len(podList.Items)
		}).WithPolling(1 * time.Second).WithTimeout(300 * time.Second).Should(BeNumerically("==", 1))

		// polling until pod.status.phase of daemonset became Running with timeout 15 sec
		podName := podList.Items[0].Name
		Eventually(func() v1.PodPhase {
			pod, err := e2e.BackgroundContext().ClientSet.CoreV1().Pods(e2e.BackgroundContext().Namespace).Get(context.TODO(), podName, metav1.GetOptions{})
			if err != nil {
				return v1.PodUnknown
			}
			return pod.Status.Phase
		}).WithPolling(1 * time.Second).WithTimeout(150 * time.Second).Should(Equal(v1.PodRunning))

		// polling the same node for resource name and quantity verification
		Eventually(func() int {
			node, err := e2e.BackgroundContext().ClientSet.CoreV1().Nodes().Get(context.TODO(), nodeName, metav1.GetOptions{})
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
		}).WithPolling(1 * time.Second).WithTimeout(600 * time.Second).Should(BeNumerically(">=", 1))
	}
}

func deployVerificationPodAndVerifyEnv(resourceName string) func() {
	return func() {
		// deploy verification pod
		_, err := e2e.BackgroundContext().ClientSet.CoreV1().Pods(e2e.BackgroundContext().Namespace).Create(context.TODO(), genVerificationPodManifest("1", resourceName), metav1.CreateOptions{})
		Expect(err).To(BeNil())

		// polling until pod.status.phase became succeeded with timeout 30 sec
		Eventually(func() v1.PodPhase {
			pod, err := e2e.BackgroundContext().ClientSet.CoreV1().Pods(e2e.BackgroundContext().Namespace).Get(context.TODO(), "verification-pod", metav1.GetOptions{})
			if err != nil {
				return v1.PodUnknown
			}
			return pod.Status.Phase
		}).WithPolling(1 * time.Second).WithTimeout(30 * time.Second).Should(Equal(v1.PodSucceeded))

		// parse allocated npu list through CoreV1().Pods().GetLogs() api
		request := e2e.BackgroundContext().ClientSet.CoreV1().Pods(e2e.BackgroundContext().Namespace).GetLogs("verification-pod", &v1.PodLogOptions{})
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

func cleanUpVerificationPodIfExist() func() {
	return func() {
		// delete verification pod if exist
		_ = e2e.BackgroundContext().ClientSet.CoreV1().Pods(e2e.BackgroundContext().Namespace).Delete(context.TODO(), "verification-pod", metav1.DeleteOptions{})
	}
}

/*func verifyInferenceEnv(resourceName string) func() {
	return func() {
		// deploy inference pod
		_, err := e2e.BackgroundContext().ClientSet.CoreV1().Pods(e2e.BackgroundContext().Namespace).Create(context.TODO(), genInferencePodManifest(resourceName), metav1.CreateOptions{})
		Expect(err).To(BeNil())

		// polling until pod.status.phase became succeeded with timeout up to 5 min since image size is bigger than 5GB
		Eventually(func() v1.PodPhase {
			pod, err := e2e.BackgroundContext().ClientSet.CoreV1().Pods(e2e.BackgroundContext().Namespace).Get(context.TODO(), "inference-pod", metav1.GetOptions{})
			if err != nil {
				return v1.PodUnknown
			}
			return pod.Status.Phase
		}).WithPolling(1 * time.Second).WithTimeout(300 * time.Second).Should(Equal(v1.PodSucceeded))
	}
}*/

/*func cleanUpInferencePod() func() {
	return func() {
		// delete inference pod
		err := e2e.BackgroundContext().ClientSet.CoreV1().Pods(e2e.BackgroundContext().Namespace).Delete(context.TODO(), "inference-pod", metav1.DeleteOptions{})
		Expect(err).To(BeNil())
	}
}*/
