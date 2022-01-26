# File: /Makefile
# Project: patch-operator
# File Created: 16-10-2021 13:14:09
# Author: Clay Risser
# -----
# Last Modified: 26-01-2022 09:21:54
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

include mkpm.mk
ifneq (,$(MKPM_READY))
include $(MKPM)/gnu

.PHONY: of-% build generate manifests install uninstall start
build: of-build
start: of-run
generate: of-generate
install: of-install
manifests: generate of-manifests
uninstall: of-uninstall
of-%:
	@$(MAKE) -s -f ./operator-framework.mk $(subst of-,,$@)

.PHONY: docker-%
docker-%:
	@$(MAKE) -s -C docker $(subst docker-,,$@)

endif
