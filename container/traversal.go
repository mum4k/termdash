package container

// traversal.go provides functions that navigate the container tree.

// rootCont returns the root container.
func rootCont(c *Container) *Container {
	for p := c.parent; p != nil; p = c.parent {
		c = p
	}
	return c
}

// visitFunc is executed during traversals when node is visited.
// If the visit function returns an error, the traversal terminates and the
// errStr is set to the text of the returned error.
type visitFunc func(*Container) error

// preOrder performs pre-order DFS traversal on the container tree.
func preOrder(c *Container, errStr *string, visit visitFunc) {
	if c == nil || *errStr != "" {
		return
	}

	if err := visit(c); err != nil {
		*errStr = err.Error()
		return
	}
	preOrder(c.first, errStr, visit)
	preOrder(c.second, errStr, visit)
}

// postOrder performs post-order DFS traversal on the container tree.
func postOrder(c *Container, errStr *string, visit visitFunc) {
	if c == nil || *errStr != "" {
		return
	}

	postOrder(c.first, errStr, visit)
	postOrder(c.second, errStr, visit)
	if err := visit(c); err != nil {
		*errStr = err.Error()
		return
	}
}
