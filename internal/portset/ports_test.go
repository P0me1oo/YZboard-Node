package portset

import (
	"reflect"
	"testing"
)

func TestParseNormalizesPortSets(t *testing.T) {
	ranges, err := Parse(" 20100,20000-20010,20005-20020,20021,443,443 ")
	if err != nil {
		t.Fatal(err)
	}
	want := []Range{{443, 443}, {20000, 20021}, {20100, 20100}}
	if !reflect.DeepEqual(ranges, want) {
		t.Fatalf("得到 %v，期望 %v", ranges, want)
	}
	whole, err := Parse("1-65535")
	if err != nil || len(whole) != 1 {
		t.Fatalf("完整端口范围不应展开：%v, %v", whole, err)
	}
}

func TestParseRejectsInvalidPortSets(t *testing.T) {
	for _, raw := range []string{"", "443,", ",443", "0", "65536", "200-100", "1-2-3", "1:3", "+443", "1;id", "-1", "1.5"} {
		if _, err := Parse(raw); err == nil {
			t.Errorf("应拒绝端口表达式 %q", raw)
		}
	}
}

func TestWithoutListenerPreservesOtherPorts(t *testing.T) {
	input := []Range{{440, 445}, {500, 500}}
	want := []Range{{440, 442}, {444, 445}, {500, 500}}
	if got := Without(input, 443); !reflect.DeepEqual(got, want) {
		t.Fatalf("得到 %v，期望 %v", got, want)
	}
	if input[0].To != 445 {
		t.Fatal("不能修改调用方的端口范围")
	}
}
