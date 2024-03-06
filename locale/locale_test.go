package locale

import (
	"testing"
)

func TestToLanguage(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want I18nLanguage
	}{
		{
			"not match",
			args{
				"zh-Hans",
			},
			ZhCN,
		},
		{
			"zh-CN",
			args{
				"zh-CN",
			},
			ZhCN,
		},
		{
			"en-US",
			args{
				"en-US",
			},
			EnUS,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToLanguage(tt.args.str); got != tt.want {
				t.Errorf("ToLanguage() = %v, want %v", got, tt.want)
			}
		})
	}
}
