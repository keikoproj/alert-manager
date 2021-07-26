package wavefront_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestWavefront(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Wavefront Suite")
}
