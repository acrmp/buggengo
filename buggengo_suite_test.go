package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func TestBuggengo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "buggengo Suite")
}

var binaryPath string

var _ = BeforeSuite(func() {
	var err error
	binaryPath, err = gexec.Build("github.com/acrmp/buggengo")
	Expect(err).ToNot(HaveOccurred())
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})
