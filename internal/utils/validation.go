package utils

import "net/url"

func IsValidURL(str string) bool {
	_, err := url.Parse(str)
	return err == nil
}

func If[T any](cond bool, vtrue, vfalse T) T {
	if cond {
		return vtrue
	}
	return vfalse
}
