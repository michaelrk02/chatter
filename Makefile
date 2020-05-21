GO = go
BUILD = $(GO) build
REPO = github.com/michaelrk02/chatter
OUT = -o build

all: proto server client

.PHONY : all

proto:
	protoc --go_out=plugins=grpc:. chatter.proto

.PHONY : proto

server: proto
	$(BUILD) $(OUT)/server $(REPO)/server

.PHONY : server

client: proto
	$(BUILD) $(OUT)/client $(REPO)/client

.PHONY : client

