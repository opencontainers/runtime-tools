#!/bin/sh

test_description='Test generate template'

. ./sharness.sh

test_expect_success CAT,ECHO,HEAD,JQ 'Test oci-runtime-tool generate --template with an empty template' "
	echo '{}' >template &&
	oci-runtime-tool generate --template template | jq . >actual &&
	cat <<-EOF >expected &&
		{
		  \"ociVersion\": \"1.0.0\",
		  \"process\": {
		    \"user\": {
		      \"uid\": 0,
		      \"gid\": 0
		    },
		    \"args\": null,
		    \"cwd\": \"/\"
		  },
		  \"root\": {
		    \"path\": \"rootfs\"
		  }
		}
	EOF
	test_cmp expected actual
"

test_expect_success CAT,HEAD,JQ 'Test oci-runtime-tool generate --template with a different version' "
	echo '{\"ociVersion\": \"1.0.0-rc9\"}' >template &&
	oci-runtime-tool generate --template template | jq . >actual &&
	cat <<-EOF >expected &&
		{
		  \"ociVersion\": \"1.0.0-rc9\",
		  \"process\": {
		    \"user\": {
		      \"uid\": 0,
		      \"gid\": 0
		    },
		    \"args\": null,
		    \"cwd\": \"/\"
		  },
		  \"root\": {
		    \"path\": \"rootfs\"
		  }
		}
	EOF
	test_cmp expected actual
"

test_done
