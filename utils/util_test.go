package utils

import (
	"fmt"
	"testing"
)

func TestCreateCaptcha(t *testing.T) {
	for i := 0; i < 10; i++ {
		got := CreateCaptcha(6)
		fmt.Printf("%v ----- %v  \n", i, got)
	}

}

func TestCreateUid(t *testing.T) {

	for i := 0; i < 5; i++ {
		got := CreateUid()
		fmt.Println("got...", got)
	}
}
