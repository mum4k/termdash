package attrrange

import (
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func TestForPosition(t *testing.T) {
	tests := []struct {
		desc string
		// if not nil, called before calling ForPosition.
		// Can add ranges.
		update        func(*Tracker) error
		pos           int
		want          *AttrRange
		wantErr       error
		wantUpdateErr bool
	}{
		{
			desc:    "fails when no ranges given",
			pos:     0,
			wantErr: ErrNotFound,
		},
		{
			desc: "fails to add a duplicate",
			update: func(tr *Tracker) error {
				if err := tr.Add(2, 5, 40); err != nil {
					return err
				}
				return tr.Add(2, 3, 41)
			},
			wantUpdateErr: true,
		},
		{
			desc: "fails when multiple given ranges, position falls before them",
			update: func(tr *Tracker) error {
				if err := tr.Add(2, 5, 40); err != nil {
					return err
				}
				return tr.Add(5, 10, 41)
			},
			pos:     1,
			wantErr: ErrNotFound,
		},
		{
			desc: "multiple given options, position falls on the lower",
			update: func(tr *Tracker) error {
				if err := tr.Add(2, 5, 40); err != nil {
					return err
				}
				return tr.Add(5, 10, 41)
			},
			pos:  2,
			want: newAttrRange(2, 5, 40),
		},
		{
			desc: "multiple given options, position falls between them",
			update: func(tr *Tracker) error {
				if err := tr.Add(2, 5, 40); err != nil {
					return err
				}
				return tr.Add(5, 10, 41)
			},
			pos:  4,
			want: newAttrRange(2, 5, 40),
		},
		{
			desc: "multiple given options, position falls on the higher",
			update: func(tr *Tracker) error {
				if err := tr.Add(2, 5, 40); err != nil {
					return err
				}
				return tr.Add(5, 10, 41)
			},
			pos:  5,
			want: newAttrRange(5, 10, 41),
		},
		{
			desc: "multiple given options, position falls after them",
			update: func(tr *Tracker) error {
				if err := tr.Add(2, 5, 40); err != nil {
					return err
				}
				return tr.Add(5, 10, 41)
			},
			pos:     10,
			wantErr: ErrNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			tr := NewTracker()
			if tc.update != nil {
				err := tc.update(tr)
				if (err != nil) != tc.wantUpdateErr {
					t.Errorf("tc.update => unexpected error:%v, wantUpdateErr:%v", err, tc.wantUpdateErr)
				}
				if err != nil {
					return
				}
			}

			got, err := tr.ForPosition(tc.pos)
			if err != tc.wantErr {
				t.Errorf("ForPosition => unexpected error:%v, wantErr:%v", err, tc.wantErr)
			}
			if err != nil {
				return
			}

			if diff := pretty.Compare(tc.want, got); diff != "" {
				t.Errorf("ForPosition => unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}
