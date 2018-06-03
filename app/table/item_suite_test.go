package table_test

import (
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/slice-d/genzai/app/table"
)

func TestItem(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Item Suite")
}

var _ = Describe("Book", func() {
	var (
	//longBook  Book
	//shortBook Book
	)

	BeforeEach(func() {
		//longBook = Book{
		//	Title:  "Les Miserables",
		//	Author: "Victor Hugo",
		//	Pages:  1488,
		//}
		//
		//shortBook = Book{
		//	Title:  "Fox In Socks",
		//	Author: "Dr. Seuss",
		//	Pages:  24,
		//}
	})

	Describe("Categorizing book length", func() {
		Context("With more than 300 pages", func() {
			It("should be a novel", func() {
				Expect("NOVEL").To(Equal("NOVEL"))
			})
		})

		Context("With fewer than 300 pages", func() {
			It("should be a short story", func() {
				Expect("SHORT STORY").To(Equal("SHORT STORY"))
			})
		})
	})
})

var _ = Describe("Book", func() {
	var (
		//longBook  Book
		//shortBook Book
		set *table.Table
	)

	BeforeEach(func() {
		//set = item.NewTable()
		//longBook = Book{
		//	Title:  "Les Miserables",
		//	Author: "Victor Hugo",
		//	Pages:  1488,
		//}
		//
		//shortBook = Book{
		//	Title:  "Fox In Socks",
		//	Author: "Dr. Seuss",
		//	Pages:  24,
		//}
	})

	Describe("Categorizing book length", func() {
		Context("With more than 300 pages", func() {
			It("should be a novel", func() {
				fmt.Println(set)
				Expect("NOVEL").To(Equal("NOVEL"))
			})
		})

		Context("With fewer than 300 pages", func() {
			It("should be a short story", func() {
				Expect("SHORT STORY").To(Equal("SHORT STORY"))
			})
		})
	})
})

var _ = Measure("it should do something hard efficiently", func(b Benchmarker) {
	runtime := b.Time("runtime", func() {
		output := 17
		time.Sleep(time.Second)
		Expect(output).To(Equal(17))
	})

	Ω(runtime.Seconds()).Should(BeNumerically("<", 1.1), "SomethingHard() shouldn't take too long.")

	b.RecordValue("disk usage (in MB)", 1.0)
}, 10)