#!/bin/bash

# Check if feature name is provided
if [ -z "$1" ]; then
  echo "Usage: ./create_feature.sh <feature_name>"
  exit 1
fi

FEATURE_NAME=$1

# Create the feature folder and subfolders
mkdir -p "internal/$FEATURE_NAME"/{types,service,handlers,repository,integration,routes}

echo "Feature '$FEATURE_NAME' created with subfolders: types, service, handlers, repository, integration, routes."