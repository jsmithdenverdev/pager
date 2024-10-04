#!/bin/bash

# Define a map of branch prefixes to version increments
declare -A version_map=(
    ["feat"]="MINOR"
    ["feature"]="MINOR"
    ["bugfix"]="PATCH"
    ["fix"]="PATCH"
    ["hotfix"]="PATCH"
    ["release"]="MINOR"
    ["major"]="MAJOR"
    ["wip"]="NOOP"
    ["experimental"]="NOOP"
    ["develop"]="NOOP"
    ["dev"]="NOOP"
    ["main"]="NOOP"
    ["master"]="NOOP"
)

# Function to extract the prefix of the branch name
get_branch_prefix() {
    local branch_name="$1"
    echo "${branch_name%%/*}"
}

# Get the branch name from the GitHub Action environment variable (GITHUB_HEAD_REF)
branch_name="$1"
if [ -z "$branch_name" ]; then
    echo "Error: No branch name provided."
    exit 1
fi

# Extract the branch prefix
branch_prefix=$(get_branch_prefix "$branch_name")

# Resolve the prefix to the corresponding version increment
version_increment="${version_map[$branch_prefix]}"

# Default to NOOP if the prefix isn't recognized
if [ -z "$version_increment" ]; then
    version_increment="NOOP"
fi

# Print the result to console
echo "Branch: $branch_name"
echo "Prefix: $branch_prefix"
echo "Version Increment: $version_increment"

# Set the output for the GitHub Action
echo "version_increment=$version_increment" >> $GITHUB_ENV
