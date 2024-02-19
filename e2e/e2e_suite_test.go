package e2e

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
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
	ClientSet    clientset.Interface
	Namespace    *v1.Namespace
}

func newFramework() (*framework, error) {
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

	return &framework{
		clientConfig: config,
		ClientSet:    clientSet,
	}, nil

}

var frk *framework

// TODO(@bg): we may need to set up kubernetes cluster in e2e-test to run test for supported versions
var _ = BeforeSuite(func() {
	newFrk, err := newFramework()
	Expect(err).To(BeNil())
	frk = newFrk

	// list namespaces to ensure api-server accessibility
	list, err := frk.ClientSet.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	Expect(err).To(BeNil())
	Expect(len(list.Items)).Should(BeNumerically(">=", 1))
})

var _ = Describe("end-to-end test", func() {
	Context("test legacy strategy", func() {
		It("deploy device-plugin helm chart for legacy strategy", func() {
			// do deployment
			// ensure deployment
			// verify strategy
			// verify resource name
		})

		It("select node and calculate expected result", func() {
			// list nodes with npu taint
			// create a pod to calculate expected result for test scenarios
			// parse expected result though CoreV1().Pods().GetLogs() api
		})

		It("request NPUs for user container only", func() {
			// deploy pod
			// parse allocated npu list through CoreV1().Pods().GetLogs() api
			// compare actual and expected
			// delete test pod
		})

		It("request NPUs for init and user container", func() {
			// deploy pod
			// parse allocated npu list through CoreV1().Pods().GetLogs() api
			// compare actual and expected
			// delete test pod
		})

		It("request NPUs for multiple user container", func() {
			// deploy pod
			// parse allocated npu list through CoreV1().Pods().GetLogs() api
			// compare actual and expected
			// delete test pod
		})

		It("request NPUs more than the node has", func() {
			// deploy a pod
			// the pod should remain in pending status
			// delete test pod
		})

		It("delete helm chart", func() {
		})
	})

	Context("test generic strategy ", func() {
		It("deploy device-plugin helm chart for generic strategy", func() {
			// do deployment
			// ensure deployment
			// verify strategy
			// verify resource name
		})

		// NOTE: skip redundant tests, legacy and generic strategy is same except resource name

		It("delete helm chart", func() {
		})
	})

	// add more tests
})
