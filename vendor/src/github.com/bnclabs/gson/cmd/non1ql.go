// +build !n1ql

package main

import "fmt"

func sortn1ql(filename string) []string {
	fmt.Println("not built with n1ql")
	return nil
}
