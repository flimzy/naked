package main

import (
	"os"
	"testing"
)

func Test_countNaked(t *testing.T) {
	tests := []struct {
		name string
		// wd sets the working directory the test should be run from.
		// used to avoid the `testdata` skipping rule for actual testdata.
		wd                                           string
		filename                                     string
		src                                          any
		wantTotal, wantNaked, wantClothed, wantMixed int
	}{
		{
			name: "no funcs",
			src:  "package foo",
		},
		{
			name: "one naked return",
			src: `package foo
			func foo() (err error) {
				return
			}`,
			wantTotal: 1,
			wantNaked: 1,
		},
		{
			name: "mixed",
			src: `package foo
			func foo() (err errorr) {
				if false {
					return nil
				}
				return
			}`,
			wantTotal:   2,
			wantNaked:   1,
			wantClothed: 1,
			wantMixed:   1,
		},
		{
			name: "with if",
			src: `package foo
func foo() error {
	f, err := os.Open("asdf")
	if err != nil{
		return err
	}
	_ = f.Close()
	return nil
}`,
			wantTotal:   2,
			wantClothed: 2,
		},
		{
			name: "generated",
			src: `package foo
// GENERATED FILE DO NOT EDIT
func foo() (err error) {
	return
}`,
		},
		{
			name:     "skip file",
			wd:       "testdata",
			filename: "skip.go",
		},
		{
			name:      "from file",
			wd:        "testdata",
			filename:  "naked.go",
			wantTotal: 1,
			wantNaked: 1,
		},
		{
			name:     "skip testdata",
			filename: "testdata/naked.go",
		},
		{
			name: "func literal",
			src: `package foo
			var _ = func() (err error) {
				return
			}`,
			wantTotal: 1,
			wantNaked: 1,
		},
		{
			name: "no return value",
			src: `package foo
			func foo() {
				return
			}`,
			wantTotal: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wd != "" {
				cwd, err := os.Getwd()
				if err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() {
					_ = os.Chdir(cwd)
				})
				if err := os.Chdir(tt.wd); err != nil {
					t.Fatal(err)
				}
			}
			total, naked, clothed, mixed, err := countNaked(tt.filename, tt.src)
			if err != nil {
				t.Fatal(err)
			}
			if total != tt.wantTotal ||
				naked != tt.wantNaked ||
				clothed != tt.wantClothed ||
				mixed != tt.wantMixed {
				t.Errorf("Unexpected result. Want: %d/%d/%d/%d, got %d/%d/%d/%d", tt.wantTotal, tt.wantNaked, tt.wantClothed, tt.wantMixed, total, naked, clothed, mixed)
			}
		})
	}
}
