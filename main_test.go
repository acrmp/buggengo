package main_test

import (
	"encoding/json"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("buggengo", func() {

	Context("when a valid go repo is provided", func() {

		var (
			candidates []map[string]any
			candidate  map[string]any
		)

		BeforeEach(func() {
			cmd := exec.Command(binaryPath, "rewrite-candidates", "_examples/cli__cli.3d27e61a")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			err = json.Unmarshal(session.Out.Contents(), &candidates)
			Expect(err).ToNot(HaveOccurred())

			for _, c := range candidates {
				Expect(c).To(HaveKey("func_name"))
				if c["func_name"] == "RESTWithNext" {
					candidate = c
				}
			}
			Expect(candidate).ToNot(BeNil())
		})

		It("provides basic function metadata", func() {
			Expect(candidate).To(HaveKeyWithValue("file_path", "_examples/cli__cli.3d27e61a/api/client.go"))
			Expect(candidate).To(HaveKeyWithValue("func_name", "RESTWithNext"))
			Expect(candidate).To(HaveKeyWithValue("func_signature", "func (c Client) RESTWithNext(hostname string, method string, p string, body io.Reader, data interface{}) (string, error)"))
			Expect(candidate).To(HaveKeyWithValue("line_start", 112.0))
			Expect(candidate).To(HaveKeyWithValue("line_end", 152.0))
		})

		It("provides a placeholder function for a LLM to implement", func() {
			Expect(candidate).To(HaveKeyWithValue("func_to_write", "func (c Client) RESTWithNext(hostname string, method string, p string, body io.Reader, data interface{}) (string, error) {\n\tpanic(\"TODO: Implement this function\")\n}"))
		})

		It("provides the full source code of the file but with the placeholder function", func() {
			Expect(candidate).To(HaveKeyWithValue("file_src_code", HavePrefix("package api\n\nimport (\n\t\"context\"\n")))
			Expect(candidate).To(HaveKeyWithValue("file_src_code", ContainSubstring("func (c Client) RESTWithNext(hostname string, method string, p string, body io.Reader, data interface{}) (string, error) {\n\tpanic(\"TODO: Implement this function\")\n}")))
		})

		It("applies the placeholder function to only a single function in the returned source", func() {
			for _, c := range candidates {
				source := gbytes.BufferWithBytes([]byte((c["file_src_code"]).(string)))
				Expect(source).To(gbytes.Say("TODO: Implement this function"))
				Expect(source).ToNot(gbytes.Say("TODO: Implement this function"))
			}
		})

		It("excludes tests", func() {
			for _, c := range candidates {
				Expect(c["file_path"]).ToNot(HaveSuffix("_test.go"))
			}
		})
	})

	Context("when the strategy is not provided", func() {
		It("errors", func() {
			cmd := exec.Command(binaryPath)
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
			Expect(string(session.Err.Contents())).To(ContainSubstring(`buggengo [strategy] [repo-directory]`))
			Expect(session.ExitCode()).ToNot(Equal(0))
		})
	})

	Context("when the repo directory is not provided", func() {
		It("errors", func() {
			cmd := exec.Command(binaryPath, "rewrite-candidates")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
			Expect(string(session.Err.Contents())).To(ContainSubstring(`buggengo [strategy] [repo-directory]`))
			Expect(session.ExitCode()).ToNot(Equal(0))
		})
	})

	Context("when the strategy is not supported", func() {
		It("errors", func() {
			cmd := exec.Command(binaryPath, "cold-fusion", "_examples/cli__cli.3d27e61a")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
			Expect(string(session.Err.Contents())).To(ContainSubstring(`buggengo [strategy] [repo-directory]`))
			Expect(session.ExitCode()).ToNot(Equal(0))
		})
	})

	Context("when too many positional arguments are provided", func() {
		It("errors", func() {
			cmd := exec.Command(binaryPath, "rewrite-candidates", "_examples/cli__cli.3d27e61a", "unexpected-argument")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
			Expect(string(session.Err.Contents())).To(ContainSubstring(`buggengo [strategy] [repo-directory]`))
			Expect(session.ExitCode()).ToNot(Equal(0))
		})
	})

	Context("when the provided repo directory does not exist", func() {
		It("errors", func() {
			cmd := exec.Command(binaryPath, "rewrite-candidates", "non-existent-directory")
			session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
			Expect(session.Err.Contents()).To(ContainSubstring(`Repo directory does not exist: "non-existent-directory"`))
			Expect(session.ExitCode()).ToNot(Equal(0))
		})
	})

	Context("when a go source file in the repo is malformed", func() {

		var (
			candidates []map[string]any
			session    *gexec.Session
		)

		BeforeEach(func() {
			cmd := exec.Command(binaryPath, "rewrite-candidates", "_examples/cli__cli.3d27e61a.malformed")
			var err error
			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit(0))

			err = json.Unmarshal(session.Out.Contents(), &candidates)
			Expect(err).ToNot(HaveOccurred())
		})

		It("does not include the malformed function in the list of candidates", func() {
			Expect(candidates).ToNot(ContainElement(HaveKeyWithValue("func_name", "MalformedFunc")))
		})

		It("logs that it had difficulty parsing the malformed file", func() {
			Expect(string(session.Err.Contents())).To(ContainSubstring(`Problem processing repo directory: "_examples/cli__cli.3d27e61a.malformed"`))
			Expect(string(session.Err.Contents())).To(ContainSubstring("_examples/cli__cli.3d27e61a.malformed/api/malformed.go"))
		})

		It("still returns candidates from earlier files in the same repo", func() {
			Expect(candidates).To(ContainElement(HaveKeyWithValue("func_name", "RESTWithNext")))
		})

		It("won't be able to return funcs from the same file", func() {
			Expect(candidates).ToNot(ContainElement(HaveKeyWithValue("func_name", "SomeOtherFuncInTheSameFile")))
			Expect(candidates).ToNot(ContainElement(HaveKeyWithValue("file_path", "_examples/cli__cli.3d27e61a.malformed/api/malformed.go")))
		})

		It("still returns candidates from later files in the same repo", func() {
			Expect(candidates).To(ContainElement(HaveKeyWithValue("func_name", "SomeOtherFuncInALaterFile")))
		})
	})

})
