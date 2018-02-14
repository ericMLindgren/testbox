package challenges

type Id = string

// Challenge - Defines challenge object
type Challenge struct {
	Id          string            `json:"id"`
	Description string            `json:"description"`
	IO          map[string]string `json:"io"`
	SampleIO    string            `json:"sampleIO"`
}

func (c Challenge) GetIOSplit() ([]string, []string) {
	ins := make([]string, len(c.IO))
	outs := make([]string, len(c.IO))

	var j int
	for i, o := range c.IO {
		ins[j] = i
		outs[j] = o
		j++
	}

	return ins, outs
}

// func (c *Challenge) StringifyCases(sep string) (string, string) {
// 	inputs := make([]string, len(c.IO))
// 	outputs := make([]string, len(c.IO))

// 	log.Printf("getCases, challenge: %v\n", c)
// 	i := 0
// 	for k, v := range c.IO {
// 		inputs[i] = k
// 		outputs[i] = v
// 		i++
// 	}

// 	return joinAndAppend(inputs, sep), joinAndAppend(outputs, sep)
// }

// func joinAndAppend(sl []string, endChar string) string {
// 	return strings.Join(sl, endChar) + endChar
// }
