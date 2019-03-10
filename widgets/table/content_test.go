package table

func ExampleContent() {
	rows := []*Row{
		NewHeader(
			NewCell("hello"),
			NewCell("world"),
		),
		NewRow(
			NewCell("1"),
			NewCell("2"),
		),
	}

	_, err := NewContent(Columns(2), rows)
	if err != nil {
		panic(err)
	}
}
