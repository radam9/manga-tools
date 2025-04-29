package ranges

import (
	"strconv"
	"strings"
)

type Range struct {
	Start float64
	End   float64
}

func Parse(rnge string) (rngs []Range, err error) {
	if rnge == "" {
		return []Range{}, nil
	}
	co := strings.Split(rnge, ",")
	var cur float64
	var begin float64
	var end float64

	for _, part := range co {
		in := strings.Split(part, "-")
		if cur, err = strconv.ParseFloat(in[0], 64); err != nil {
			return
		}
		if len(in) == 2 {
			begin = cur
			if end, err = strconv.ParseFloat(in[1], 64); err != nil {
				return
			}
		}

		if begin == 0 {
			begin = cur
		}
		if end < cur {
			end = cur
		}
		if begin > cur {
			continue
		}

		rngs = append(rngs, Range{
			Start: begin,
			End:   end,
		})

		begin = 0
		end = 0
	}

	return
}
