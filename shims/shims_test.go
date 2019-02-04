package shims_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/ansicleaner"
	"gopkg.in/jarcoal/httpmock.v1"

	"github.com/cloudfoundry/libbuildpack/shims"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -source=supplier.go --destination=mocks_shims_test.go --package=shims_test

var _ = Describe("Shims", func() {

	Describe("Supplier", func() {
		var (
			supplier        shims.Supplier
			v2AppDir        string
			v2BuildpacksDir string
			v3AppDir        string
			v2DepsDir       string
			depsIndex       string
			tempDir         string
		)

		BeforeEach(func() {
			var err error

			tempDir, err = ioutil.TempDir("", "tmp")
			Expect(err).NotTo(HaveOccurred())

			v2AppDir = filepath.Join(tempDir, "build")
			Expect(os.MkdirAll(v2AppDir, 0777)).To(Succeed())

			v3AppDir = filepath.Join(tempDir, "cnb-app")

			v2DepsDir = filepath.Join(tempDir, "deps")
			depsIndex = "0"

			v2BuildpacksDir = filepath.Join(tempDir, "buildpacks")
		})

		JustBeforeEach(func() {
			Expect(os.MkdirAll(filepath.Join(v2DepsDir, depsIndex), 0777)).To(Succeed())

			Expect(os.MkdirAll(filepath.Join(v2BuildpacksDir, depsIndex), 0777)).To(Succeed())

			supplier = shims.Supplier{
				V2AppDir:       v2AppDir,
				V2BuildpackDir: filepath.Join(v2BuildpacksDir, depsIndex),
				V3AppDir:       v3AppDir,
				V2DepsDir:      v2DepsDir,
				DepsIndex:      depsIndex,
			}
		})

		AfterEach(func() {
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		})

		Context("SetUpFirstV3Buildpack", func() {
			It("Moves V2AppDir to V3AppDir if it has not already been moved", func() {
				Expect(v3AppDir).NotTo(BeADirectory())
				Expect(supplier.SetUpFirstV3Buildpack()).To(Succeed())
				Expect(v3AppDir).To(BeADirectory())
			})

			It("Writes a sentinel file in the app dir", func() {
				Expect(supplier.SetUpFirstV3Buildpack()).To(Succeed())
				Expect(filepath.Join(v3AppDir, ".cloudfoundry", shims.SENTINEL)).To(BeAnExistingFile())
			})

			It("Writes a symlink at the V2AppDir to a fake file with a clear error message", func() {
				Expect(supplier.SetUpFirstV3Buildpack()).To(Succeed())
				linkDst, err := os.Readlink(v2AppDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(linkDst).To(Equal(shims.ERROR_FILE))
			})

			It("Does nothing if V2AppDir has already been moved", func() {
				Expect(os.Remove(v2AppDir)).To(Succeed())
				Expect(os.Symlink("some-file", v2AppDir)).To(Succeed())
				Expect(supplier.SetUpFirstV3Buildpack()).To(Succeed())
			})
		})

		Context("SaveOrderToml", func() {
			It("copies its order.toml to the deps/orders dir", func() {
				Expect(ioutil.WriteFile(filepath.Join(v2BuildpacksDir, depsIndex, "order.toml"), []byte(""), 0666)).To(Succeed())
				Expect(supplier.SaveOrderToml()).To(Succeed())
				Expect(filepath.Join(v2DepsDir, "order", "order"+depsIndex+".toml")).To(BeAnExistingFile())
			})
		})
	})

	Describe("Finalizer", func() {
		var (
			finalizer    shims.Finalizer
			mockCtrl     *gomock.Controller
			mockDetector *MockDetector
			tempDir,
			v2AppDir,
			v3AppDir,
			v2DepsDir,
			v3LayersDir,
			v3BuildpacksDir,
			planMetadata,
			groupMetadata,
			profileDir,
			binDir,
			depsIndex    string
		)

		BeforeEach(func() {
			var err error

			mockCtrl = gomock.NewController(GinkgoT()) // TODO DOn't use GInkgoT?
			mockDetector = NewMockDetector(mockCtrl)

			depsIndex = "0"

			tempDir, err = ioutil.TempDir("", "tmp")
			Expect(err).NotTo(HaveOccurred())

			v2AppDir = filepath.Join(tempDir, "v2_app")
			Expect(os.MkdirAll(v2AppDir, 0777)).To(Succeed())

			v3AppDir = filepath.Join(tempDir, "v3_app")
			Expect(os.MkdirAll(v3AppDir, 0777)).To(Succeed())

			v2DepsDir = filepath.Join(tempDir, "deps")

			v3LayersDir = filepath.Join(tempDir, "layers")
			Expect(os.MkdirAll(v3LayersDir, 0777)).To(Succeed())

			v3BuildpacksDir = filepath.Join(tempDir, "cnbs")
			Expect(os.MkdirAll(v3BuildpacksDir, 0777)).To(Succeed())

			planMetadata = filepath.Join(tempDir, "plan.toml")
			groupMetadata = filepath.Join(tempDir, "group.toml")

			profileDir = filepath.Join(tempDir, "profile")
			Expect(os.MkdirAll(profileDir, 0777)).To(Succeed())

			binDir = filepath.Join(tempDir, "bin")
			Expect(os.MkdirAll(binDir, 0777)).To(Succeed())

			Expect(os.Setenv("CF_STACK", "some-stack")).To(Succeed())

		})

		JustBeforeEach(func() {
			Expect(os.MkdirAll(filepath.Join(v2DepsDir, depsIndex), 0777)).To(Succeed())

			finalizer = shims.Finalizer{
				V2AppDir:        v2AppDir,
				V3AppDir:        v3AppDir,
				V2DepsDir:       v2DepsDir,
				V3LayersDir:     v3LayersDir,
				V3BuildpacksDir: v3BuildpacksDir,
				DepsIndex:       depsIndex,
				PlanMetadata:    planMetadata,
				GroupMetadata:   groupMetadata,
				ProfileDir:      profileDir,
				BinDir:          binDir,
				Detector:        mockDetector,
			}
		})

		AfterEach(func() {
			mockCtrl.Finish()
			Expect(os.Unsetenv("CF_STACK")).To(Succeed())
			Expect(os.RemoveAll(tempDir)).To(Succeed())
		})

		Context("RunV3Detect", func() {
			It("runs detection when group or plan metadata does not exist", func() {
				mockDetector.
					EXPECT().
					Detect()
				Expect(finalizer.RunV3Detect()).To(Succeed())
			})

			It("does NOT run detection when group and plan metadata exists", func() {
				Expect(ioutil.WriteFile(groupMetadata, []byte(""), 0666)).To(Succeed())
				Expect(ioutil.WriteFile(planMetadata, []byte(""), 0666)).To(Succeed())

				mockDetector.
					EXPECT().
					Detect().
					Times(0)
				Expect(finalizer.RunV3Detect()).To(Succeed())
			})
		})

		Context("MoveV3Layers", func() {
			BeforeEach(func() {
				Expect(os.MkdirAll(filepath.Join(v3LayersDir, "config"), 0777)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(v3LayersDir, "config", "metadata.toml"), []byte(""), 0666)).To(Succeed())

				Expect(os.MkdirAll(filepath.Join(v3LayersDir, "layer"), 0777)).To(Succeed())
				Expect(os.MkdirAll(filepath.Join(v3LayersDir, "anotherLayer"), 0777)).To(Succeed())
			})

			It("moves the layers to deps dir and metadata to build dir", func() {
				Expect(finalizer.MoveV3Layers()).To(Succeed())
				Expect(filepath.Join(v2AppDir, ".cloudfoundry", "metadata.toml")).To(BeAnExistingFile())
				Expect(filepath.Join(v2DepsDir, "layer")).To(BeAnExistingFile())
				Expect(filepath.Join(v2DepsDir, "anotherLayer")).To(BeAnExistingFile())
			})

		})

		Context("IncludePreviousV2Buildpacks", func() {
			var (
				createDirs, createFiles []string
			)

			BeforeEach(func() {
				depsIndex = "2"
				createDirs = []string{"bin", "lib"}
				createFiles = []string{"config.yml"}
				for _, dir := range createDirs {
					Expect(os.MkdirAll(filepath.Join(v2DepsDir, "0", dir), 0777)).To(Succeed())
				}

				for _, file := range createFiles {
					Expect(ioutil.WriteFile(filepath.Join(v2DepsDir, "0", file), []byte(file), 0666)).To(Succeed())
				}

				Expect(ioutil.WriteFile(groupMetadata, []byte(""), 0666)).To(Succeed())
				Expect(ioutil.WriteFile(planMetadata, []byte(""), 0666)).To(Succeed())
			})

			It("copies v2 layers and metadata where v3 lifecycle expects them for build and launch", func() {
				By("not failing if a layer has already been moved")
				Expect(finalizer.IncludePreviousV2Buildpacks()).To(Succeed())

				By("putting the v2 layers in the corrent directory structure")
				for _, dir := range createDirs {
					Expect(filepath.Join(v3LayersDir, "buildpack.0", "layer", dir)).To(BeADirectory())
				}

				for _, file := range createFiles {
					Expect(filepath.Join(v3LayersDir, "buildpack.0", "layer", file)).To(BeAnExistingFile())
				}

				By("writing the group metadata in the order the buildpacks should be sourced")
				groupMetadataContents, err := ioutil.ReadFile(groupMetadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(groupMetadataContents)).To(Equal(`[[buildpacks]]
  id = "buildpack.0"
  version = ""

[[buildpacks]]
  id = "buildpack.1"
  version = ""
`))
			})
		})

		Context("MoveV2Layers", func() {
			It("moves directories and creates the dst dir if it doesn't exist", func() {
				Expect(finalizer.MoveV2Layers(filepath.Join(v2DepsDir, depsIndex), filepath.Join(v3LayersDir, "buildpack.0", "layers.0"))).To(Succeed())
				Expect(filepath.Join(v3LayersDir, "buildpack.0", "layers.0")).To(BeADirectory())
			})
		})

		Context("RenameEnvDir", func() {
			It("renames the env dir to env.build", func() {
				Expect(os.Mkdir(filepath.Join(v3LayersDir, "env"), 0777)).To(Succeed())
				Expect(finalizer.RenameEnvDir(v3LayersDir)).To(Succeed())
				Expect(filepath.Join(v3LayersDir, "env.build")).To(BeADirectory())
			})

			It("does nothing when the env dir does NOT exist", func() {
				Expect(finalizer.RenameEnvDir(v3LayersDir)).To(Succeed())
				Expect(filepath.Join(v3LayersDir, "env.build")).NotTo(BeADirectory())
			})
		})

		Context("UpdateGroupTOML", func() {
			BeforeEach(func() {
				depsIndex = "1"
				Expect(ioutil.WriteFile(groupMetadata, []byte(`[[buildpacks]]
  id = "org.cloudfoundry.buildpacks.nodejs"
  version = "0.0.2"
[[buildpacks]]
  id = "org.cloudfoundry.buildpacks.npm"
  version = "0.0.3"`), 0777)).To(Succeed())
			})

			It("adds v2 buildpacks to the group.toml", func() {
				Expect(finalizer.UpdateGroupTOML("buildpack.0")).To(Succeed())
				groupMetadataContents, err := ioutil.ReadFile(groupMetadata)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(groupMetadataContents)).To(Equal(`[[buildpacks]]
  id = "buildpack.0"
  version = ""

[[buildpacks]]
  id = "org.cloudfoundry.buildpacks.nodejs"
  version = "0.0.2"

[[buildpacks]]
  id = "org.cloudfoundry.buildpacks.npm"
  version = "0.0.3"
`))
			})
		})

		Context("AddFakeCNBBuildpack", func() {
			It("adds the v2 buildpack as a no-op cnb buildpack", func() {
				Expect(os.Setenv("CF_STACK", "cflinuxfs3")).To(Succeed())
				Expect(finalizer.AddFakeCNBBuildpack("buildpack.0")).To(Succeed())
				buildpackTOML, err := ioutil.ReadFile(filepath.Join(v3BuildpacksDir, "buildpack.0", "latest", "buildpack.toml"))
				Expect(err).ToNot(HaveOccurred())
				Expect(string(buildpackTOML)).To(Equal(`[buildpack]
  id = "buildpack.0"
  name = "buildpack.0"
  version = "latest"

[[stacks]]
  id = "org.cloudfoundry.stacks.cflinuxfs3"
`))

				Expect(filepath.Join(v3BuildpacksDir, "buildpack.0", "latest", "bin", "build")).To(BeAnExistingFile())
			})
		})

		Context("MergeOrderTOMLs with unique buildpacks", func() {
			BeforeEach(func() {
				const (
					ORDER1 = "order1.toml"
					ORDER2 = "order2.toml"
				)

				orderPath := filepath.Join(v2DepsDir, "order")
				Expect(os.MkdirAll(orderPath, 0777)).To(Succeed())
				orderFileA := filepath.Join(orderPath, ORDER1)
				Expect(ioutil.WriteFile(orderFileA, []byte(`[[groups]]
  labels = ["testA"]

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpA"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpB"
    version = "latest"
	optional = true`), 0777)).To(Succeed())

				orderFileB := filepath.Join(orderPath, ORDER2)
				Expect(ioutil.WriteFile(orderFileB, []byte(`[[groups]]
  labels = ["testA"]

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpC"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpD"
    version = "latest"`), 0777)).To(Succeed())
			})

			It("merges the order files", func() {
				Expect(finalizer.MergeOrderTOMLs()).To(Succeed())
				orderTOML, err := ioutil.ReadFile(filepath.Join(v2DepsDir, "order.toml"))

				Expect(err).ToNot(HaveOccurred())
				Expect(string(orderTOML)).To(ContainSubstring(`[[groups]]
  labels = ["testA"]

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpA"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpB"
    version = "latest"
    optional = true

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpC"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpD"
    version = "latest"`))
			})
		})

		Context("MergeOrderTOMLs with duplicate buildpacks", func() {
			BeforeEach(func() {
				const (
					ORDER1 = "order1.toml"
					ORDER2 = "order2.toml"
				)

				orderPath := filepath.Join(v2DepsDir, "order")
				Expect(os.MkdirAll(orderPath, 0777)).To(Succeed())
				orderFileA := filepath.Join(orderPath, ORDER1)
				Expect(ioutil.WriteFile(orderFileA, []byte(`[[groups]]
  labels = ["testA"]

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpA"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpB"
    version = "latest"`), 0777)).To(Succeed())

				orderFileB := filepath.Join(orderPath, ORDER2)
				Expect(ioutil.WriteFile(orderFileB, []byte(`[[groups]]
  labels = ["testA"]

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpA"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpC"
    version = "latest"`), 0777)).To(Succeed())
			})

			It("merges the order files", func() {
				Expect(finalizer.MergeOrderTOMLs()).To(Succeed())
				orderTOML, err := ioutil.ReadFile(filepath.Join(v2DepsDir, "order.toml"))

				Expect(err).ToNot(HaveOccurred())
				Expect(string(orderTOML)).To(ContainSubstring(`[[groups]]
  labels = ["testA"]

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpA"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpB"
    version = "latest"

  [[groups.buildpacks]]
    id = "this.is.a.fake.bpC"
    version = "latest"`))
			})
		})
	})

	Describe("Releaser", func() {
		var (
			releaser shims.Releaser
			v2AppDir string
			buf      *bytes.Buffer
		)

		BeforeEach(func() {
			var err error

			v2AppDir, err = ioutil.TempDir("", "build")
			Expect(err).NotTo(HaveOccurred())

			contents := `
			buildpacks = ["some.buildpacks", "some.other.buildpack"]
			[[processes]]
			type = "web"
			command = "npm start"
			`
			Expect(os.MkdirAll(filepath.Join(v2AppDir, ".cloudfoundry"), 0777)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(v2AppDir, ".cloudfoundry", "metadata.toml"), []byte(contents), 0666)).To(Succeed())

			buf = &bytes.Buffer{}

			releaser = shims.Releaser{
				MetadataPath: filepath.Join(v2AppDir, ".cloudfoundry", "metadata.toml"),
				Writer:       buf,
			}
		})

		AfterEach(func() {
			Expect(os.RemoveAll(v2AppDir)).To(Succeed())
		})

		It("runs with the correct arguments and moves things to the correct place", func() {
			Expect(releaser.Release()).To(Succeed())
			Expect(buf.Bytes()).To(Equal([]byte("default_process_types:\n  web: npm start\n")))
			Expect(filepath.Join(v2AppDir, ".cloudfoundry", "metadata.toml")).NotTo(BeAnExistingFile())
		})
	})

	Describe("CNBInstaller", func() {
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
