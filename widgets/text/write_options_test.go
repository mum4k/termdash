package text

import (
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/mum4k/termdash/cell"
)

func TestGivenWOpts(t *testing.T) {
	tests := []struct {
		desc  string
		given givenWOpts
		pos   int
		want  *optsRange
	}{
		{
			desc:  "no write options given results in defaults",
			given: nil,
			pos:   1,
			want: &optsRange{
				low:  0,
				high: 0,
				opts: &writeOptions{
					cellOpts: &cell.Options{},
				},
			},
		},
		{
			desc: "multiple given options, position falls before them",
			given: givenWOpts{
				2: &optsRange{
					low:  2,
					high: 5,
					opts: &writeOptions{
						cellOpts: &cell.Options{
							FgColor: cell.ColorBlue,
						},
					},
				},
				5: &optsRange{
					low:  5,
					high: 10,
					opts: &writeOptions{
						cellOpts: &cell.Options{
							FgColor: cell.ColorRed,
						},
					},
				},
			},
			pos: 1,
			want: &optsRange{
				low:  0,
				high: 0,
				opts: &writeOptions{
					cellOpts: &cell.Options{},
				},
			},
		},
		{
			desc: "multiple given options, position falls on the lower",
			given: givenWOpts{
				2: &optsRange{
					low:  2,
					high: 5,
					opts: &writeOptions{
						cellOpts: &cell.Options{
							FgColor: cell.ColorBlue,
						},
					},
				},
				5: &optsRange{
					low:  5,
					high: 10,
					opts: &writeOptions{
						cellOpts: &cell.Options{
							FgColor: cell.ColorRed,
						},
					},
				},
			},
			pos: 2,
			want: &optsRange{
				low:  2,
				high: 5,
				opts: &writeOptions{
					cellOpts: &cell.Options{
						FgColor: cell.ColorBlue,
					},
				},
			},
		},
		{
			desc: "multiple given options, position falls between them",
			given: givenWOpts{
				2: &optsRange{
					low:  2,
					high: 5,
					opts: &writeOptions{
						cellOpts: &cell.Options{
							FgColor: cell.ColorBlue,
						},
					},
				},
				5: &optsRange{
					low:  5,
					high: 10,
					opts: &writeOptions{
						cellOpts: &cell.Options{
							FgColor: cell.ColorRed,
						},
					},
				},
			},
			pos: 4,
			want: &optsRange{
				low:  2,
				high: 5,
				opts: &writeOptions{
					cellOpts: &cell.Options{
						FgColor: cell.ColorBlue,
					},
				},
			},
		},
		{
			desc: "multiple given options, position falls on the higher",
			given: givenWOpts{
				2: &optsRange{
					low:  2,
					high: 5,
					opts: &writeOptions{
						cellOpts: &cell.Options{
							FgColor: cell.ColorBlue,
						},
					},
				},
				5: &optsRange{
					low:  5,
					high: 10,
					opts: &writeOptions{
						cellOpts: &cell.Options{
							FgColor: cell.ColorRed,
						},
					},
				},
			},
			pos: 5,
			want: &optsRange{
				low:  5,
				high: 10,
				opts: &writeOptions{
					cellOpts: &cell.Options{
						FgColor: cell.ColorRed,
					},
				},
			},
		},
		{
			desc: "multiple given options, position falls after them",
			given: givenWOpts{
				2: &optsRange{
					low:  2,
					high: 5,
					opts: &writeOptions{
						cellOpts: &cell.Options{
							FgColor: cell.ColorBlue,
						},
					},
				},
				5: &optsRange{
					low:  5,
					high: 10,
					opts: &writeOptions{
						cellOpts: &cell.Options{
							FgColor: cell.ColorRed,
						},
					},
				},
			},
			pos: 10,
			want: &optsRange{
				low:  0,
				high: 0,
				opts: &writeOptions{
					cellOpts: &cell.Options{},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			got := tc.given.forPosition(tc.pos)
			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("forPosition => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}
