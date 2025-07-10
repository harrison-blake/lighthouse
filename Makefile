.PHONY: all
all: install test build sync invalidate-cache

.PHONY: clean
clean:
	rm -rf public

.PHONY: test
test:
	go test

.PHONY: install
install: 
	go install .

.PHONY: sync
sync: check-target-dir check-s3-bucket
	aws s3 sync $(TARGET_DIR) s3://$(S3_BUCKET)

.PHONY: build
build: 
	lighthouse

.PHONY: invalidate-cache
invalidate-cache: check-cloudfront-id-sub check-cloudfront-id-root
	aws cloudfront create-invalidation --distribution-id $(CLOUDFRONT_ID_SUB) --paths /
	aws cloudfront create-invalidation --distribution-id $(CLOUDFRONT_ID_ROOT) --paths /

.PHONY: check-cloudfront-id-sub
check-cloudfront-id-sub:
ifndef CLOUDFRONT_ID_SUB
	$(error CLOUDFRONT_ID_SUB is required)
endif

.PHONY: check-cloudfront-id-root
check-cloudfront-id-root:
ifndef CLOUDFRONT_ID_ROOT
	$(error CLOUDFRONT_ID_ROOT is required)
endif

.PHONY: check-target-dir
check-target-dir:
ifndef TARGET_DIR
	$(error TARGET_DIR is required)
endif

.PHONY: check-s3-bucket
check-s3-bucket:
ifndef S3_BUCKET
	$(error S3_BUCKET is required)
endif
