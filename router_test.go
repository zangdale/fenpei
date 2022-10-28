package main

import "testing"

func TestCheck(t *testing.T) {
	root := NewBaseRouter()
	root.AddRouter("/1/2/3/4/5/6", nil)
	root.AddRouter("/1/2/3/4/5", nil)
	root.AddRouter("/2/1/5/6", nil)
	t.Log(root.CheckRouter("/1/2/3"))
	t.Log(root.CheckRouter("/"))
	t.Log(root.CheckRouter("/2/1/5/6"))
	t.Log(root.CheckRouter("/1/2/3/4/5/6"))
}
