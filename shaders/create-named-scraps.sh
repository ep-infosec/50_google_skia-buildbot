#!/bin/bash

# Create the named SkSL scraps.
#
# The scrapexchange server should be available locally, such as forwarding the
# the production protected endpoint to localhost:
#
#    $ kubectl port-forward service/scrapexchange 9000
#
# Note that this is in place of an Admin UI for the Scrap Exchange server. Once
# that is complete it should be the canconical way to create named scraps.

# Create a name for each scrap.
curl --silent -X PUT -d "{\"Hash\": \"8faf0a93848175d382bc8feaef584339b75a911aa7a9d8a47834d4b05c801a08\", \"Description\": \"Shader Inputs\"}" -H 'Content-Type: application/json' http://localhost:9000/_/names/sksl/@inputs
curl --silent -X PUT -d "{\"Hash\": \"3124d06f75fa4a5138784e80592909d094a50dfd7a53aed82f660c1c021fa628\", \"Description\": \"Shader Inputs\"}" -H 'Content-Type: application/json' http://localhost:9000/_/names/sksl/@iResolution
curl --silent -X PUT -d "{\"Hash\": \"12f6c3ecf3f26b1b734d0f254b5bb97cb7d395a9e4829fbab9cda3fef9e3ad9e\", \"Description\": \"Shader Inputs\"}" -H 'Content-Type: application/json' http://localhost:9000/_/names/sksl/@iTime
curl --silent -X PUT -d "{\"Hash\": \"229312432fbd5c54b471d766a0da29cd7d22f53b7a9a90f3b014a87114d02dd1\", \"Description\": \"Shader Inputs\"}" -H 'Content-Type: application/json' http://localhost:9000/_/names/sksl/@iMouse
curl --silent -X PUT -d "{\"Hash\": \"f1be7449cdd7b15ab39efa681f5191bbaa55b1c58f582cb16c55186dd95a24e0\", \"Description\": \"Shader Inputs\"}" -H 'Content-Type: application/json' http://localhost:9000/_/names/sksl/@default
curl --silent -X PUT -d "{\"Hash\": \"498ba63d9e91c960fd7d242ba4d9e68ce16fbf1396aac2820f005a8464daac94\", \"Description\": \"Shader Inputs\"}" -H 'Content-Type: application/json' http://localhost:9000/_/names/sksl/@iImage
curl --silent -X PUT -d "{\"Hash\": \"498ba63d9e91c960fd7d242ba4d9e68ce16fbf1396aac2820f005a8464daac94\", \"Description\": \"Shader Inputs\"}" -H 'Content-Type: application/json' http://localhost:9000/_/names/sksl/@defaultChildShader


# List all named sksl scraps.
curl --silent http://localhost:9000/_/names/sksl/
