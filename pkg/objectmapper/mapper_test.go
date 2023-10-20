package objectmapper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type Nested struct {
	A *string       `json:"a"`
	B float32       `json:"b"`
	C MoreNesting   `json:"c"`
	F []MoreNesting `json:"f"`
}

type MoreNesting struct {
	D []int  `json:"d"`
	E string `json:"e"`
}

type StructSrc struct {
	N Nested `json:"n"`
}

type Dest struct {
	A string                 `json:"a" om:"n.a"`
	B int                    `json:"b" om:"n.b,32"`
	D *string                `json:"d" om:"n.c.d[1]"`
	E int                    `json:"e" om:"n.c.e"`
	F []int                  `json:"f" om:"n.c.d[:]"`
	G MoreNesting            `json:"g" om:"n.f[0]"`
	H map[string]interface{} `json:"h" om:"n.f[1]"`
	I string                 `json:"i" om:"n.f[0].e"`
	J uint                   `json:"j" om:"n.f[0].d[1]"`
	K bool                   `json:"k"`
}

type Dest2 struct {
	A string                 `json:"a" customTag:"n.a"`
	B int                    `json:"b" customTag:"n.b;32"`
	D *string                `json:"d" customTag:"n.c.d[invalid];pj"`
	E int                    `json:"e" customTag:"n.c.e"`
	F []int                  `json:"f" customTag:"n.c.d[1:]"`
	G MoreNesting            `json:"g" customTag:"n.f[100]"`
	H map[string]interface{} `json:"h" customTag:"n.f[1]"`
	I string                 `json:"i" customTag:"n.f[-1].e"`
	J float32                `json:"j" customTag:"n.f[0].d[1]"`
	K []bool                 `json:"k" customTag:";true"`
	L StructSrc              `json:"l" customTag:"."`
}

func TestTemp(t *testing.T) {
	str := "abc"
	str2 := "32"
	str3 := "pj"

	s := StructSrc{
		N: Nested{
			A: &str,
			B: 123.2,
			C: MoreNesting{D: []int{23, 32}},
			F: []MoreNesting{{D: []int{2, 3}, E: "xyz"}, {D: []int{4, 5}, E: "pqr"}},
		},
	}

	tests := []struct {
		name        string
		source      interface{}
		expectedErr error
		exec        func()
	}{
		{
			name:   "master test case 1",
			source: s,
			exec: func() {
				om := ObjectMapper{}
				d := Dest{}
				expectedRes := Dest{
					A: "abc",
					B: 32,
					D: &str2,
					F: []int{23, 32},
					G: MoreNesting{D: []int{2, 3}, E: "xyz"},
					H: map[string]any{"d": []any{float64(4), float64(5)}, "e": "pqr"},
					I: "xyz",
					J: 3,
				}
				assert.Nil(t, om.Map(s, &d))
				assert.Equal(t, expectedRes, d)
			},
		},
		{
			name:   "master test case 2",
			source: s,
			exec: func() {
				om := ObjectMapper{Tag: "customTag", DefaultValueSeparator: ";"}
				d := Dest2{}
				expectedRes := Dest2{
					A: "abc",
					B: 32,
					D: &str3,
					F: []int{32},
					G: MoreNesting{D: []int{4, 5}, E: "pqr"},
					H: map[string]any{"d": []any{float64(4), float64(5)}, "e": "pqr"},
					I: "xyz",
					J: 3,
					K: []bool{true},
					L: s,
				}
				assert.Nil(t, om.Map(s, &d))
				assert.Equal(t, expectedRes, d)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(_ *testing.T) {
			test.exec()
		})
	}
}
