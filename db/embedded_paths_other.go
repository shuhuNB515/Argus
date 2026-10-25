//go:build !windows

package db

func prepareEmbeddedRoot(root string) (string, error) { return root, nil }
