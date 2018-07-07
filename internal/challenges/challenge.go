package challenges

import "fmt"

// ID is a string to identify challenges
type ID = int64

// Challenge - Defines challenge object
type Challenge struct {
	ID        ID       `json:"id"`
	Name      string   `json:"name"`
	ShortDesc string   `json:"shortDesc"`
	LongDesc  string   `json:"longDesc"`
	Tags      tagList  `json:"tags"`
	SampleIO  CaseList `json:"sampleIO"`
	Cases     CaseList `json:"cases"`
}

func (c Challenge) String() string {
	return fmt.Sprintf("ID: %d, Name: %s, ShortDesct: OMIT, LongDesc: OMIT, Tags: OMIT, Cases: %s, SampleIO: %s", c.ID, c.Name, c.Cases, c.SampleIO)
}

type TestCase struct {
	Input  string `json:"input"`
	Expect string `json:"expect"`
	Desc   string `json:"desc,omitempty"`
}

func (t TestCase) String() string {
	str := fmt.Sprintf("<TestCase> Input: %s Expect: %s ", t.Input, t.Expect)
	if t.Desc != "" {
		str += fmt.Sprintf("Desc: %s", t.Desc)
	}
	return str
}

type tagList []string
type CaseList []TestCase

func dummyChallenge() *Challenge {
	return &Challenge{
		Name:      "Echo Stdin",
		ShortDesc: "Simply echo stdin to stdout",
		LongDesc:  "",
		Tags:      tagList{"trivial"},
		SampleIO:  CaseList{TestCase{"hello world", "hello world", ""}},
		Cases:     CaseList{TestCase{"hello", "hello", ""}, TestCase{"123!?3", "123!?3", ""}, TestCase{", !.fe", ", !.fe", ""}},
	}
}

// GetIOSplit splits the IO map of in -> out into two separate slices
func (c Challenge) GetIOSplit() ([]string, []string) {
	ins := make([]string, len(c.Cases))
	outs := make([]string, len(c.Cases))

	for i, c := range c.Cases {
		ins[i] = c.Input
		outs[i] = c.Expect
	}

	return ins, outs
}
