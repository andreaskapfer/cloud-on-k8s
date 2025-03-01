# Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
# or more contributor license agreements. Licensed under the Elastic License 2.0;
# you may not use this file except in compliance with the Elastic License 2.0.

# This Makefile is mostly used for continuous integration.

ROOT_DIR               := $(CURDIR)/..
GO_MOUNT_PATH          ?= /go/src/github.com/elastic/cloud-on-k8s
DOCKER_LOGIN           ?= eckadmin

-include $(ROOT_DIR)/.env

vault := ../hack/retry.sh 5 vault

# This is set to avoid the issue described in https://github.com/hashicorp/vault/issues/6710
VAULT_CLIENT_TIMEOUT = 120
VAULT_ROOT_PATH ?= secret/devops-ci/cloud-on-k8s

# BUILD_ID is present during run on Jenkins machine, but not on dev box, hence using it here to distinguish between those cases
ifndef VAULT_TOKEN
ifdef BUILD_ID
VAULT_TOKEN = $(shell $(vault) write -address=$(VAULT_ADDR) -field=token auth/approle/login role_id=$(VAULT_ROLE_ID) secret_id=$(VAULT_SECRET_ID))
DOCKERHUB_LOGIN = $(shell VAULT_TOKEN=$(VAULT_TOKEN) $(vault) read -address=$(VAULT_ADDR) -field=username secret/release/docker-hub-eck)
DOCKERHUB_PASSWORD = $(shell VAULT_TOKEN=$(VAULT_TOKEN) $(vault) read -address=$(VAULT_ADDR) -field=token secret/release/docker-hub-eck)
else
VAULT_TOKEN = $(shell $(vault) write -address=$(VAULT_ADDR) -field=token auth/github/login token=$(GITHUB_TOKEN))
# we use roleId as a string that has to be there for authn/z for CI, but it's empty and not needed for local execution
NOT_USED := $(shell test -e $(ROOT_DIR)/deployer-config.yml && sed "s;roleId:.*;token: $(GITHUB_TOKEN);g" $(ROOT_DIR)/deployer-config.yml > tmp && mv tmp $(ROOT_DIR)/deployer-config.yml)
endif
endif

CI_SHA1  := $(shell md5sum $(ROOT_DIR)/go.mod $(ROOT_DIR)/.ci/Dockerfile | awk '{print $$1}' | md5sum | awk '{print $$1}' | cut -c-8)
CI_IMAGE ?= docker.elastic.co/eck-ci/eck-ci-tools:$(CI_SHA1)

# volume used to share files between CI container and OCP3 Ansible container
ECK_CI_VOLUME := eck-ci-home-$(shell date '+%N')

print-ci-image:
	@ echo $(CI_IMAGE)

# runs $TARGET in context of CI container and dev makefile
ci:
	@ $(MAKE) DOCKER_CMD="make $(TARGET)" ci-internal

ci-interactive:
	@ $(MAKE) DOCKER_OPTS=-i DOCKER_CMD=bash ci-internal

ci-internal: ci-build-image
	# --userns=host to support Docker daemon host configured to run containers only in user namespaces
	@ docker volume create --name $(ECK_CI_VOLUME) > /dev/null && \
	  docker run --userns=host --rm -t $(DOCKER_OPTS) \
		-v /var/run/docker.sock:/var/run/docker.sock \
		-v $(ECK_CI_VOLUME):/root/ \
		-v $(ROOT_DIR):$(GO_MOUNT_PATH) \
		-e SHARED_VOLUME_NAME=$(ECK_CI_VOLUME) -e ENABLE_FIPS \
		-e VAULT_ADDR -e VAULT_TOKEN \
		-w $(GO_MOUNT_PATH) \
		--network="host" \
		$(CI_IMAGE) \
		bash -c "$(DOCKER_CMD)" ; exit=$$?; \
	  docker volume rm $(ECK_CI_VOLUME) > /dev/null; exit $$exit

# build and push the CI image only if it does not yet exist
ci-build-image: DOCKER_PASSWORD = $(shell VAULT_TOKEN=$(VAULT_TOKEN) $(vault) read -address=$(VAULT_ADDR) -field=value $(VAULT_ROOT_PATH)/eckadmin)
ci-build-image: write-ci-docker-creds
	@ docker pull -q $(CI_IMAGE) | grep -v -E 'Downloading|Extracting|Verifying|complete' || \
	( \
		../hack/retry.sh 5 docker build \
			-f $(ROOT_DIR)/.ci/Dockerfile \
			-t $(CI_IMAGE) \
			--label "commit.hash=$(shell git rev-parse --short --verify HEAD)" \
			$(ROOT_DIR) && \
		../hack/docker.sh -l -p $(CI_IMAGE) \
	)

# make Docker creds available from inside the CI container through the .registry.env file
write-ci-docker-creds:
	@ echo "DOCKER_LOGIN=$(DOCKER_LOGIN)"              > ${ROOT_DIR}/.registry.env
	@ echo "DOCKER_PASSWORD=$(DOCKER_PASSWORD)"       >> ${ROOT_DIR}/.registry.env
	@ echo "DOCKERHUB_LOGIN=$(DOCKERHUB_LOGIN)"       >> ${ROOT_DIR}/.registry.env
	@ echo "DOCKERHUB_PASSWORD=$(DOCKERHUB_PASSWORD)" >> ${ROOT_DIR}/.registry.env

##  Test

# all artifacts needed to run e2e tests
get-test-artifacts: monitoring-secrets.json test-license.json license.key

ifneq (,$(IS_SNAPSHOT_BUILD))
       SECRET_FIELD_PREFIX ?= dev-
get-test-artifacts: dev-private.key
endif

# read a dev private key used to generate dev licenses against non-production builds
dev-private.key:
	@ VAULT_TOKEN=$(VAULT_TOKEN) $(vault) read -address=$(VAULT_ADDR) -field=dev-privkey  ${VAULT_ROOT_PATH}/license | base64 --decode > dev-private.key

# read some test licenses from Vault for E2E license tests
test-license.json:
	@ VAULT_TOKEN=$(VAULT_TOKEN) $(vault) read -address=$(VAULT_ADDR) -field=$(SECRET_FIELD_PREFIX)enterprise  ${VAULT_ROOT_PATH}/test-licenses > test-license.json

# read connection info and credentials to the E2E tests monitoring Elasticsearch cluster, to be used during E2E tests
monitoring-secrets.json:
	@ VAULT_TOKEN=$(VAULT_TOKEN) $(vault) read -address=$(VAULT_ADDR) -field=data -format=json ${VAULT_ROOT_PATH}/monitoring-cluster > monitoring-secrets.json

##  Build

# read Elastic public key from Vault into license.key, to build the operator for E2E tests or for a release
license.key:
	@ VAULT_TOKEN=$(VAULT_TOKEN) $(vault) read -address=$(VAULT_ADDR) -field=$(SECRET_FIELD_PREFIX)pubkey ${VAULT_ROOT_PATH}/license | base64 --decode > license.key

##  Release

# read AWS creds from Vault for YAML upload to S3
yaml-upload: VAULT_AWS_CREDS = secret/cloud-team/cloud-ci/eck-release
yaml-upload: AWS_ACCESS_KEY_ID = $(shell VAULT_TOKEN=$(VAULT_TOKEN) $(vault) read -address=$(VAULT_ADDR) -field=access-key-id $(VAULT_AWS_CREDS))
yaml-upload: AWS_SECRET_ACCESS_KEY = $(shell VAULT_TOKEN=$(VAULT_TOKEN) $(vault) read -address=$(VAULT_ADDR) -field=secret-access-key $(VAULT_AWS_CREDS))
yaml-upload:
ifndef IMG_VERSION
	$(error IMG_VERSION not set to upload YAML to S3)
endif
	@ $(MAKE) \
		DOCKER_OPTS="-e VERSION=$(IMG_VERSION) -e AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) -e AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY)" \
		DOCKER_CMD="hack/manifest-upload.sh" \
		ci-internal
