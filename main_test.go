package main

import "testing"

func Test_getURL(t *testing.T) {
	t.Parallel()
	type args struct {
		line string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Google",
			args: args{
				line: "google.com\n",
			},
			want: "https://www.google.com",
		}, {
			name: "Yandex",
			args: args{
				line: "yandex.ru\n",
			},
			want: "https://www.yandex.ru",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := getURL(tt.args.line); got != tt.want {
				t.Errorf("getURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
