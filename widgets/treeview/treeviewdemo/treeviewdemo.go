// treeviewdemo.go
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/donut"
	"github.com/mum4k/termdash/widgets/sparkline"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/mum4k/termdash/widgets/treeview"
)

// NodeData holds arbitrary data associated with a tree node.
type NodeData struct {
	// Label is the label of the node.
	Label string
	// PID is the process ID associated with the node.
	PID int
	// CPUUsage is a slice of CPU usage percentages.
	CPUUsage []int
	// MemoryUsage is the current memory usage percentage.
	MemoryUsage int
	// LastCPUUsage is the last recorded CPU usage percentage.
	LastCPUUsage int
	// LastMemoryUsage is the last recorded memory usage percentage.
	LastMemoryUsage int
}

// Helper function to generate smoother data.
func generateNextValue(prev int) int {
	delta := rand.Intn(5) - 2 // Random change between -2 and +2
	next := prev + delta
	if next < 0 {
		next = 0
	} else if next > 100 {
		next = 100
	}
	return next
}

// fetchStaticData creates a static tree structure with associated data.
func fetchStaticData() ([]*treeview.TreeNode, map[string]*NodeData, error) {
	var nodeDataMap = make(map[string]*NodeData)

	// Seed random number generator.
	rand.Seed(time.Now().UnixNano())

	// Create the root node.
	root := &treeview.TreeNode{
		Label: "Applications",
	}

	// Helper function to recursively build the tree and assign IDs.
	var buildTree func(tn *treeview.TreeNode, path string)
	buildTree = func(tn *treeview.TreeNode, path string) {
		tn.ID = generateNodeID(path, tn.Label)

		if len(tn.Children) == 0 {
			// Leaf node: assign data.
			pid := rand.Intn(9000) + 1000 // Random PID between 1000 and 9999.
			initialCPUUsage := rand.Intn(100)
			initialMemoryUsage := rand.Intn(100)
			data := &NodeData{
				Label:           tn.Label,
				PID:             pid,
				CPUUsage:        []int{},
				MemoryUsage:     initialMemoryUsage,
				LastCPUUsage:    initialCPUUsage,
				LastMemoryUsage: initialMemoryUsage,
			}
			// Initialize CPUUsage slice with initial values.
			for i := 0; i < 20; i++ {
				data.LastCPUUsage = generateNextValue(data.LastCPUUsage)
				data.CPUUsage = append(data.CPUUsage, data.LastCPUUsage)
			}
			nodeDataMap[tn.ID] = data
		} else {
			// Recursively assign IDs to child nodes.
			for _, child := range tn.Children {
				buildTree(child, tn.ID)
			}
		}
	}

	// Build the tree structure.
	for i := 1; i <= 10; i++ {
		appNode := &treeview.TreeNode{
			Label: fmt.Sprintf("Application %d", i),
		}
		for j := 1; j <= 5; j++ {
			subAppNode := &treeview.TreeNode{
				Label: fmt.Sprintf("SubApp %d.%d", i, j),
			}
			for k := 1; k <= 3; k++ {
				featureNode := &treeview.TreeNode{
					Label: fmt.Sprintf("Feature %d.%d.%d", i, j, k),
				}
				subAppNode.Children = append(subAppNode.Children, featureNode)
			}
			appNode.Children = append(appNode.Children, subAppNode)
		}
		root.Children = append(root.Children, appNode)
	}

	// Assign IDs and build nodeDataMap.
	buildTree(root, "")

	return []*treeview.TreeNode{root}, nodeDataMap, nil
}

// generateNodeID creates a consistent node ID.
func generateNodeID(path string, label string) string {
	if path == "" {
		return label
	}
	return fmt.Sprintf("%s/%s", path, label)
}

// updateWidgets updates the widgets with data from the selected node.
func updateWidgets(data *NodeData, memDonut *donut.Donut, spark *sparkline.SparkLine, detailText *text.Text) {
	// Use data to update widgets.
	spark.Add([]int{data.LastCPUUsage})
	memDonut.Percent(data.MemoryUsage)

	detailText.Reset()
	detailText.Write(fmt.Sprintf(
		"Selected Node: %s\n"+
			"PID:          %d\n"+
			"CPU Usage:    %d%%\n"+
			"Memory Usage: %d%%\n",
		data.Label,
		data.PID,
		data.LastCPUUsage,
		data.MemoryUsage,
	))
}

func main() {
	// Initialize terminal.
	t, err := tcell.New()
	if err != nil {
		log.Fatalf("failed to create terminal: %v", err)
	}
	defer t.Close()

	// Fetch static tree data.
	processTree, nodeDataMap, err := fetchStaticData()
	if err != nil {
		log.Fatalf("failed to fetch static data: %v", err)
	}

	// Create widgets.
	memDonut, err := donut.New(
		donut.CellOpts(cell.FgColor(cell.ColorGreen)),
		donut.Label("Memory Usage", cell.FgColor(cell.ColorYellow)),
	)
	if err != nil {
		log.Fatalf("failed to create donut widget: %v", err)
	}

	spark, err := sparkline.New(
		sparkline.Color(cell.ColorBlue),
		sparkline.Label("CPU Usage"),
	)
	if err != nil {
		log.Fatalf("failed to create sparkline widget: %v", err)
	}

	detailText, err := text.New(
		text.WrapAtWords(),
	)
	if err != nil {
		log.Fatalf("failed to create text widget: %v", err)
	}

	// Mutex to protect access to the selected node.
	var mu sync.Mutex
	var selectedNodeID string

	// Create TreeView widget with logging enabled for debugging.
	tv, err := treeview.New(
		treeview.Label("Applications TreeView"),
		treeview.Nodes(processTree...),
		treeview.LabelColor(cell.ColorBlue),
		treeview.CollapsedIcon("▶"),
		treeview.ExpandedIcon("▼"),
		treeview.WaitingIcons([]string{"◐", "◓", "◑", "◒"}),
		treeview.LeafIcon(""),
		treeview.Indentation(2),
		treeview.Truncate(true),
		treeview.EnableLogging(false),
	)
	if err != nil {
		log.Fatalf("failed to create TreeView: %v", err)
	}

	// Assign OnClick handlers to leaf nodes only.
	var assignOnClick func(tn *treeview.TreeNode)
	assignOnClick = func(tn *treeview.TreeNode) {
		// Assign OnClick only to leaf nodes.
		if len(tn.Children) == 0 {
			tn := tn // Capture range variable.
			tn.OnClick = func() error {
				mu.Lock()
				selectedNodeID = tn.ID
				mu.Unlock()

				data := nodeDataMap[tn.ID]
				if data != nil {
					updateWidgets(data, memDonut, spark, detailText)
				} else {
					// Clear the widgets if no data is associated.
					spark.Add([]int{0})
					memDonut.Percent(0)
					detailText.Reset()
					detailText.Write(fmt.Sprintf("Selected Node: %s\n", tn.Label))
					detailText.Write("No data available for this node.")
				}

				// Simulate a longer-running process to make the spinner visible.
				time.Sleep(500 * time.Millisecond)

				return nil
			}
		}
		for _, child := range tn.Children {
			assignOnClick(child)
		}
	}

	for _, tn := range processTree {
		assignOnClick(tn)
	}

	// Build grid layout.
	builder := grid.New()
	builder.Add(
		grid.RowHeightPerc(70,
			grid.ColWidthPerc(40,
				grid.Widget(tv,
					container.Border(linestyle.Light),
					container.BorderTitle("Applications TreeView"),
					container.Focused(), // Set initial focus to the TreeView
				),
			),
			grid.ColWidthPerc(30,
				grid.Widget(memDonut,
					container.Border(linestyle.Light),
					container.BorderTitle("Memory Usage"),
				),
			),
			grid.ColWidthPerc(30,
				grid.Widget(spark,
					container.Border(linestyle.Light),
					container.BorderTitle("CPU Sparkline"),
				),
			),
		),
		grid.RowHeightPerc(30,
			grid.Widget(detailText,
				container.Border(linestyle.Light),
				container.BorderTitle("Process Details"),
			),
		),
	)

	gridOpts, err := builder.Build()
	if err != nil {
		log.Fatalf("failed to build grid layout: %v", err)
	}

	// Create container.
	c, err := container.New(
		t,
		gridOpts...,
	)
	if err != nil {
		log.Fatalf("failed to create container: %v", err)
	}

	// Context for termdash.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start a goroutine to continuously update the widgets.
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				mu.Lock()
				id := selectedNodeID
				mu.Unlock()

				if id != "" {
					data := nodeDataMap[id]
					if data != nil {
						// Update data with new values.
						data.LastCPUUsage = generateNextValue(data.LastCPUUsage)
						data.CPUUsage = append(data.CPUUsage, data.LastCPUUsage)
						if len(data.CPUUsage) > 20 {
							data.CPUUsage = data.CPUUsage[1:]
						}
						data.LastMemoryUsage = generateNextValue(data.LastMemoryUsage)
						data.MemoryUsage = data.LastMemoryUsage

						// Update widgets.
						updateWidgets(data, memDonut, spark, detailText)
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Global key press handler to exit on 'q' or 'Esc'.
	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == keyboard.KeyEsc || k.Key == 'q' {
			cancel()
		}
	}

	// Run termdash.
	if err := termdash.Run(ctx, t, c,
		termdash.KeyboardSubscriber(quitter),
		termdash.RedrawInterval(500*time.Millisecond),
	); err != nil {
		log.Fatalf("failed to run termdash: %v", err)
	}

	// Ensure spinner ticker is stopped.
	tv.StopSpinnerTicker()
}
