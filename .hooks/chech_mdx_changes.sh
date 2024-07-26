#!/usr/bin/env bash
changed_files=$(git diff --cached --name-only)

mdx_files=()

for file in $changed_files; do
    if [[ $file =~ (tasks\.json|definition\.json|setup\.json)$ ]]; then
        component_dir=$(dirname "$file" | sed -E 's|/config$||')
        mdx_file="$component_dir/README.mdx"
        if [[ ! " ${mdx_files[@]} " =~ " ${mdx_file} " ]]; then
            mdx_files+=("$mdx_file")
        fi
    fi
done

need_to_update=false
for mdx_file in "${mdx_files[@]}"; do
    if ! echo "$changed_files" | grep -q "$mdx_file"; then
        need_to_update=true
        echo "$mdx_file is not updated."
    fi
done

if $need_to_update; then
    echo " Running make build-doc && gen-doc..."
    make build-doc
    make gen-doc
fi

exit 0
