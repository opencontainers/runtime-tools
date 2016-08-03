#!/bin/sh

test_description='Test generate output'

. ./sharness.sh

test_expect_success CAT,HEAD 'Test oci-runtime-tool generate writing to stdout' "
	oci-runtime-tool generate | head -n2 >actual &&
	cat <<EOF >expected &&
{
	\"ociVersion\": \"1.0.0\",
EOF
	test_cmp expected actual
"

test_expect_success CAT,HEAD 'Test oci-runtime-tool generate --output' "
	oci-runtime-tool generate --output config.json &&
	head -n2 config.json >actual &&
	cat <<EOF >expected &&
{
	\"ociVersion\": \"1.0.0\",
EOF
	test_cmp expected actual
"

test_done
