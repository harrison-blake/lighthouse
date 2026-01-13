.PHONY: deploy
deploy: install test build sync invalidate-cache-all

.PHONY: clean
clean:
	rm -rf public

.PHONY: install
install: 
	go install .

.PHONY: test
test:
	go test ./...

.PHONY: build
build: 
	lighthouse build

.PHONY: sync
sync: check-s3-bucket
	aws s3 sync ./public s3://$(S3_BUCKET)

.PHONY: invalidate-cache-all
invalidate-cache: check-cloudfront-id-sub check-cloudfront-id-root
	aws cloudfront create-invalidation --distribution-id $(CLOUDFRONT_ID_SUB) --paths /*
	aws cloudfront create-invalidation --distribution-id $(CLOUDFRONT_ID_ROOT) --paths /*

.PHONY: invalidate-cache-about
invalidate-cache-about: check-cloudfront-id-sub check-cloudfront-id-root
	aws cloudfront create-invalidation --distribution-id $(CLOUDFRONT_ID_SUB) --paths /about
	aws cloudfront create-invalidation --distribution-id $(CLOUDFRONT_ID_ROOT) --paths /about

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

.PHONY: check-s3-bucket
check-s3-bucket:
ifndef S3_BUCKET
	$(error S3_BUCKET is required)
endif
