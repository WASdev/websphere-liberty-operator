#!/bin/bash

readonly build_script_dir="$(dirname "$0")"

echo "Downloading build scripts"

branch="main"
build_script_folder_url="https://api.github.ibm.com/repos/websphere/operators/contents/scripts/build?ref=${branch}"
build_scripts=($(curl -H "Authorization: token ${GITHUB_ACCESS_TOKEN}" -H "Accept: application/vnd.github.v3+json" ${build_script_folder_url} | jq -r ".[] | .name"))

for build_script in "${build_scripts[@]}"; do
    build_script_url="https://api.github.ibm.com/repos/websphere/operators/contents/scripts/build/${build_script}?ref=${branch}"
    curl -H "Authorization: token ${GITHUB_ACCESS_TOKEN}" -H "Accept: application/vnd.github.v3+json" "${build_script_url}" | jq -r ".content" | base64 --decode > ${build_script_dir}/${build_script}
    chmod +x ${build_script_dir}/${build_script}
done
