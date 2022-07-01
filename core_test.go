package core

import (
	"fmt"
	"os"
	"testing"
)

func TestImage(t *testing.T) {
	s := &ImageSt{file: "img/a.gif"}
	dst, err := s.Scale(240, 240, 80)
	dir, _ := os.Getwd()

	fmt.Println(dst, err, dir)

	dst, err = s.Clip(0, 0, 1024, 1024, 90)
	fmt.Println(dst, err, dir)
}
