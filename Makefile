# File: /Makefile
# Project: patch-operator
# File Created: 16-10-2021 13:14:09
# Author: Clay Risser
# -----
# Last Modified: 16-10-2021 13:28:32
# Modified By: Clay Risser
# -----
# BitSpur Inc (c) Copyright 2021
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

IMAGE_TAG_BASE := registry.gitlab.com/bitspur/community/patch-operator

include mkpm.mk
ifneq (,$(MKPM))
-include $(MKPM)/gnu

.DEFAULT_GOAL := build

.PHONY: of-%
of-%:
	@$(MAKE) -s -f ./operator-framework.mk $(subst of-,,$@)

.PHONY: build docker-build generate manifests
build: of-build
docker-build: of-docker-build
generate: of-generate
manifests: generate of-manifests

endif
