package table

func Example_DisplayDataOnly() {
	rows := []*Row{
		NewRow(
			NewCell("hello"),
			NewCell("world"),
		),
		NewRow(
			NewCell("hello"),
			NewCell("world"),
		),
	}

	_, err := NewContent(Columns(3), rows)
	if err != nil {
		panic(err)
	}

	data := [][]string{}
	myRows := []*Row{}
	for _, dataRow := range data {
		cells := []*Cell{}
		for _, dataCol := range dataRow {
			cells = append(cells, NewCell(dataCol))
		}
		myRows = append(myRows, NewRow(cells...))
	}
}

func Example_ColorInheritance() {

}

func Example_ColAndRowSpan() {

}

func Example_TextTrimmingAndWrapping() {

}

func Example_AddRow() {

}

func Example_DeleteRow() {

}

func Example_Callback() {

}
