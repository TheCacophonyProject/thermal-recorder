package main

type TestCamera struct {
}

func (cam *TestCamera) ResX() int {
	return 160
}
func (cam *TestCamera) ResY() int {
	return 120
}
func (cam *TestCamera) FPS() int {
	return 9
}
