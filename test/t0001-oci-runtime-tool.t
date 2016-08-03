#!/bin/sh

test_description='Test ocitools global options'

. ./sharness.sh

test_expect_success CAT,HEAD 'Test oci-runtime-tool --help' "
	oci-runtime-tool --help | head -n2 >actual &&
	cat <<-EOF >expected &&
		NAME:
		   oci-runtime-tool - OCI (Open Container Initiative) runtime tools
	EOF
	test_cmp expected actual
"

test_expect_success ECHO,SED 'Test oci-runtime-tool --version' "
	oci-runtime-tool --version | sed 's/commit: [0-9a-f]*$/commit: HASH/' >actual &&
	echo 'oci-runtime-tool version 0.0.1, commit: HASH' >expected &&
	test_cmp expected actual
"

test_done
