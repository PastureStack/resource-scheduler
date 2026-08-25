module github.com/PastureStack/resource-scheduler

go 1.26.0

toolchain go1.27.0

require (
	github.com/go-viper/mapstructure/v2 v2.5.0
	github.com/rancher/event-subscriber v0.0.0-20170216231139-9a4724dc5dfe
	github.com/rancher/go-rancher v0.1.1-0.20161220063330-2c43ff300f3e
	github.com/sirupsen/logrus v1.10.1
	github.com/urfave/cli/v3 v3.11.0
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c
)

require (
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/kr/pretty v0.2.1 // indirect
	github.com/kr/text v0.1.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
)

replace github.com/rancher/event-subscriber => ./third_party/event-subscriber

replace github.com/rancher/go-rancher => ./third_party/go-rancher
