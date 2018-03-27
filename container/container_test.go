package container

// Example demonstrates how to use the Container API.
func Example() {
	New( // Create the root container.
		/* terminal = */ nil,
		SplitHorizontal(),
	).First( // This is the top half part of the terminal.
		SplitVertical(),
	).First( // Left side on the top.
		VerticalAlignTop(),
		PlaceWidget( /* widget = */ nil),
	).Parent().Second( // Right side on the top.
		HorizontalAlignRight(),
		PlaceWidget( /* widget = */ nil),
	).Parent().Parent().Second( // Bottom half of the terminal.
		PlaceWidget( /* widget = */ nil),
	)
}
