package server

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "API Suite")
}

var _ = Describe("Web", func() {
	port := 9500
	var server *Web

	BeforeSuite(func() {
		// Start a server up
		server = NewWeb(fmt.Sprintf(":%d", port))
		err := server.Start()
		Expect(err).NotTo(HaveOccurred())
		Expect(server.Addr()).ShouldNot(BeNil())
		Expect(server.Addr().Port).Should(Equal(port))
	})

	AfterSuite(func() {
		// Shut the server down
		Expect(server.Stop()).To(BeNil())
	})

	Describe("Starting the server with a specified port", func() {
		Context("when the port requested is currently in use", func() {
			It("should let the OS pick a free port and automatically use it without failing", func() {
				s := NewWeb(fmt.Sprintf(":%d", port))
				err := s.Start()
				Expect(err).To(BeNil())
				addr := s.Addr()
				Expect(addr).ShouldNot(BeNil())
				if addr != nil {
					fmt.Println(fmt.Sprintf("OS assigned port %d", addr.Port))
					Expect(addr.Port).ShouldNot(Equal(port))
				}
				Expect(s.Stop()).To(BeNil())
			})
		})
	})

	//Describe("Listing all browsers", func() {
	//	Context("when at least Chrome is installed locally", func() {
	//		It("should have at least Chrome in the reply", func() {
	//			reply, err := server.ListBrowsers(context.Background(), &api.ListBrowsersRequest{})
	//			Expect(err).NotTo(HaveOccurred())
	//			Expect(reply).ShouldNot(BeNil())
	//		})
	//	})
	//})
})

//var _ = Measure("ListBrowsers", func(b Benchmarker) {
//	s := NewWeb(":9001")
//	err := s.Start()
//	Expect(err).Should(BeNil())
//
//	b.RecordValue("pending", 1)
//
//	s.Stop()
//}, 10)
