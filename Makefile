BUILD_NUMBER := $(shell date +'%Y%m%d%H%M')
PLATFORM_FILES="$(PWD)/main/main.go"
LDFLAGS ?= $(LDFLAGS:)
export GOBIN = $(PWD)/build/aws_appconfig
GO=go

run-package: ##同步包管理
	GOPROXY=https://goproxy.cn go mod vendor
	GOPROXY=https://goproxy.cn go mod tidy


run-server: ## 启动服务器
	@echo Running aws-appconfig-server for development
	$(GO) run -mod=vendor -ldflags '$(LDFLAGS)' $(PLATFORM_FILES) --name=aws_appconfig >> console.log 2>&1 &


build-linux: ## 编译linux安装包
	@echo 编译安装环境
	rm -rf ./build
	mkdir -p ./build
	env GOOS=linux GOARCH=amd64 $(GO) build -o $(GOBIN) -ldflags '$(LDFLAGS)' $(PLATFORM_FILES)

build-docker: build-linux ## 编译Docker镜像
	@echo "Build Docker"
	@echo "$(BUILD_NUMBER)"
	cp ./build/aws_appconfig ./docker/aws_appconfig
	docker build -f ./docker/Dockerfile ./docker -t aws_appconfig:$(BUILD_NUMBER)
	rm ./docker/aws_appconfig
	docker tag aws_appconfig:$(BUILD_NUMBER) neuxs.duxinglangzi.com/aws_appconfig:$(BUILD_NUMBER)
	## docker push neuxs.duxinglangzi.com/aws_appconfig:$(BUILD_NUMBER)


help: ## 帮助
	@grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' ./Makefile | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
