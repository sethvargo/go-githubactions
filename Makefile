# Copyright 2020 The Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

export GO111MODULE = on
export GOFLAGS = -mod=vendor
export CGO_ENABLED = 0

deps:
	@go get -mod="" -u -t ./...
	@go mod tidy
	@go mod vendor
.PHONY: deps

test:
	@go test -parallel=40 ./...
.PHONY: test
