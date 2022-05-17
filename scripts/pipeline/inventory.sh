#!/usr/bin/env bash

list_entries () {
  git checkout "$1" > /dev/null

  while IFS= read -r app_name; do
    echo "${app_name//.\//}"
  done <<< "$(find -P . -mindepth 1 -type f -not -path '*/\.*' -not -path '*.md' -type f)"
}

INVENTORY_PATH="$(get_env inventory-path)"
APP_REPO_PATH=$(pwd)
cd $INVENTORY_PATH
commit_hash=$1
entries=$(list_entries "$commit_hash")

#
# write all inventory entry names in a JSON array
#
entries_json="[]"

while IFS= read -r line; do
  if [ -n "$line" ]; then
    entries_json=$(echo "$entries_json" | jq -c --arg element "$line" '. + [$element]')
  fi
done <<< "${entries}"

echo -n "$entries_json" > "$WORKSPACE/inventory-entries-list.json"

printf "Inventory entries: \n%s \n" "$(jq '.' "$WORKSPACE/inventory-entries-list.json")" >&2

set_env INVENTORY_ENTRIES_PATH "inventory-entries-list.json"
export INVENTORY_ENTRIES_PATH="inventory-entries-list.json"
cd $APP_REPO_PATH