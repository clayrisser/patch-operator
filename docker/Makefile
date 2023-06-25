# File: /Makefile
# Project: docker
# File Created: 19-10-2021 07:37:12
# Author: Clay Risser
# -----
# Last Modified: 25-06-2023 14:12:49
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

include ../mkpm.mk
ifneq (,$(MKPM_READY))
include $(MKPM)/gnu

export DOCKER_COMPOSE := docker-compose
export CONTEXT := ..
export NAME := bitspur/rock8s/patch-operator
export REGISTRY := registry.gitlab.com
export VERSION := 0.1.0

include $(MKPM)/docker

endif
