package challenges

// _ "github.com/jinzhu/gorm/dialects/mysql"
// _ "github.com/jinzhu/gorm/dialects/sqlite"

// ID is a string to identify challenges
type ID = int64

// Challenge - Defines challenge object
type Challenge struct {
	// gorm.Model // takes care of ID and some other record keeping
	ID        ID       `json:"id"`
	Name      string   `json:"name"`
	ShortDesc string   `json:"shortDesc"`
	LongDesc  string   `json:"longDesc"`
	Tags      tagList  `json:"tags"`
	Cases     caseList `json:"cases"`
	SampleIO  caseList `json:"sampleIO"`
}

type testCase struct {
	Desc   string `json:"desc,omitempty"`
	Input  string `json:"input"`
	Expect string `json:"expect"`
}

type tagList []string
type caseList []testCase

func dummyChallenge() *Challenge {
	return &Challenge{
		Name:      "Dummy Challenge",
		ShortDesc: "A dummy challenge for testing purposes",
		LongDesc:  "Lorem ipsum dolor sit amet, consectetur adipisicing elit. Deleniti libero, illo laborum, minus quisquam iusto animi neque explicabo! Commodi perspiciatis hic et cum exercitationem error voluptatum doloribus itaque consequuntur nisi!",
		Tags:      tagList{"dummy", "test", "useless"},
		Cases:     caseList{testCase{"dummy", "a", "a"}},
		SampleIO:  caseList{testCase{"", "a", "a"}},
	}
}

// GetIOSplit splits the IO map of in -> out into two separate slices
func (c Challenge) GetIOSplit() ([]string, []string) {
	ins := make([]string, len(c.SampleIO))
	outs := make([]string, len(c.SampleIO))

	for i, c := range c.SampleIO {
		ins[i] = c.Input
		outs[i] = c.Expect
	}

	return ins, outs
}
