package launchsecurity

import (
	expect "github.com/google/goexpect"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	virtconfig "kubevirt.io/kubevirt/pkg/virt-config"
	"kubevirt.io/kubevirt/tests"
	"kubevirt.io/kubevirt/tests/console"
	"kubevirt.io/kubevirt/tests/framework/checks"
	"kubevirt.io/kubevirt/tests/libvmi"
)

var _ = Describe("[sig-compute]AMD Secure Encrypted Virtualization (SEV)", func() {
	BeforeEach(func() {
		checks.SkipTestIfNoFeatureGate(virtconfig.WorkloadEncryptionSEV)
		checks.SkipTestIfNotSEVCapable()
		tests.BeforeTestCleanup()
	})

	DescribeTable("When starting a SEV or SEV-ES VM",
		func(isESEnabled bool) {
			const secureBoot = false
			dmesgRet := "SEV"
			if isESEnabled {
				dmesgRet = "SEV SEV-ES"
			}
			vmi := libvmi.NewFedora(libvmi.WithUefi(secureBoot), libvmi.WithSEV(isESEnabled))
			vmi = tests.RunVMIAndExpectLaunch(vmi, 240)

			By("Expecting the VirtualMachineInstance console")
			Expect(console.LoginToFedora(vmi)).To(Succeed())

			By("Verifying that SEV is enabled in the guest")
			err := console.SafeExpectBatch(vmi, []expect.Batcher{
				&expect.BSnd{S: "\n"},
				&expect.BExp{R: console.PromptExpression},
				&expect.BSnd{S: "dmesg | grep --color=never SEV\n"},
				&expect.BExp{R: "AMD Memory Encryption Features active: " + dmesgRet},
				&expect.BSnd{S: "\n"},
				&expect.BExp{R: console.PromptExpression},
			}, 30)
			Expect(err).ToNot(HaveOccurred())
		},
		// SEV-ES disabled, SEV enabled
		Entry("It should launch with base SEV features enabled", false),
		// SEV-ES enabled
		Entry("It should launch with SEV-ES features enabled", true),
	)
})
