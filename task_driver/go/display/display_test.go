package display

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncateError(t *testing.T) {

	test := func(input, expect string) {
		assert.Equal(t, expect, truncateError(input))
	}

	// Too small to truncate.
	test("", "")
	test("abc", "abc")

	// Max number of lines.
	test(`1
2
3
4
5
6
7
8
9
10
11
12
13
14
15
16
17
18
19
20`, `1
2
3
4
5
6
7
8
9
10
11
12
13
14
15
16
17
18
19
20`)

	// Trim final newline.
	test(`1
2
3
4
5
6
7
8
9
10
11
12
13
14
15
16
17
18
19
20
`, `1
2
3
4
5
6
7
8
9
10
11
12
13
14
15
16
17
18
19
20`)

	// One line is cut off.
	test(`1
2
3
4
5
6
7
8
9
10
11
12
13
14
15
16
17
18
19
20
21`, `...2
3
4
5
6
7
8
9
10
11
12
13
14
15
16
17
18
19
20
21`)
	// Right at the line and char limit.
	test(`01abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
02abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
03abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
04abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
05abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
06abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
07abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
08abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
09abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
10abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
11abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
12abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
13abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
14abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
15abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
16abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
17abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
18abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
19abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
20abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUV`, `01abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
02abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
03abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
04abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
05abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
06abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
07abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
08abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
09abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
10abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
11abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
12abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
13abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
14abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
15abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
16abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
17abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
18abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
19abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
20abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUV`)

	// Just over the char limit.
	test(`01abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
02abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
03abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
04abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
05abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
06abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
07abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
08abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
09abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
10abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
11abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
12abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
13abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
14abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
15abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
16abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
17abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
18abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
19abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
20abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVW`, `...cdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
02abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
03abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
04abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
05abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
06abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
07abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
08abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
09abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
10abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
11abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
12abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
13abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
14abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
15abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
16abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
17abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
18abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
19abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTU
20abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVW`)
}
