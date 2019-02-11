package shims_test

import (
"bytes"
"github.com/cloudfoundry/libbuildpack"
"github.com/cloudfoundry/libbuildpack/ansicleaner"
"github.com/cloudfoundry/libbuildpack/shims"
. "github.com/onsi/ginkgo"
. "github.com/onsi/gomega"
"gopkg.in/jarcoal/httpmock.v1"
"io/ioutil"
"os"
"path/filepath"
"time"
)

var _ = Describe("Shims", func() {
	Describe("Installer", func() {
		BeforeEach(func() {
			Expect(os.Setenv("CF_STACK", "cflinuxfs3")).To(Succeed())

			httpmock.Reset()

			contents, err := ioutil.ReadFile("fixtures/bpA.tgz")
			Expect(err).ToNot(HaveOccurred())

			httpmock.RegisterResponder("GET", "https://a-fake-url.com/bpA.tgz",
				httpmock.NewStringResponder(200, string(contents)))

			contents, err = ioutil.ReadFile("fixtures/bpB.tgz")
			Expect(err).ToNot(HaveOccurred())

			httpmock.RegisterResponder("GET", "https://a-fake-url.com/bpB.tgz",
				httpmock.NewStringResponder(200, string(contents)))
		})

		AfterEach(func() {
			Expect(os.Unsetenv("CF_STACK")).To(Succeed())
		})

		It("installs the latest/unique buildpacks from an order.toml", func() {
			tmpDir, err := ioutil.TempDir("", "")
			Expect(err).ToNot(HaveOccurred())
			defer os.RemoveAll(tmpDir)

			buffer := new(bytes.Buffer)
			logger := libbuildpack.NewLogger(ansicleaner.New(buffer))

			manifest, err := libbuildpack.NewManifest("fixtures", logger, time.Now())
			Expect(err).To(BeNil())

			installer := shims.NewCNBInstaller(manifest)

			Expect(installer.InstallCNBS("fixtures/order.toml", tmpDir)).To(Succeed())
			Expect(filepath.Join(tmpDir, "this.is.a.fake.bpA", "1.0.1", "a.txt")).To(BeAnExistingFile())
			Expect(filepath.Join(tmpDir, "this.is.a.fake.bpB", "1.0.2", "b.txt")).To(BeAnExistingFile())
			Expect(filepath.Join(tmpDir, "this.is.a.fake.bpA", "latest")).To(BeAnExistingFile())
			Expect(filepath.Join(tmpDir, "this.is.a.fake.bpB", "latest")).To(BeAnExistingFile())
		})
	})
})
