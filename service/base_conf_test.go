package service_test

import (
	"ex-depth-wss/service"
	"testing"
)

func TestUniq(t *testing.T) {
	strArr := []string{"123", "abc", "123", "abc "}
	uniArr := service.Uniq(strArr)
	if len(uniArr) == 3 {
		t.Log(uniArr)
	} else {
		t.Error("failed", uniArr)
	}
}

func BenchmarkUniq(b *testing.B) {

}

