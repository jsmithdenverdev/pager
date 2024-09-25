#!/bin/bash

# Set the directory where the executables are located
CMD_DIR="./cmd"
DIST_DIR="./dist"
SEMVER=$1

# Create the dist directory if it doesn't exist
mkdir -p $DIST_DIR

# Function to build and zip an executable
build_and_zip_executable() {
  EXECUTABLE=$1

  # Create the output directory for the executable
  OUTPUT_DIR="$DIST_DIR/$EXECUTABLE"
  mkdir -p $OUTPUT_DIR

  # Build the executable and output to the dist directory as "bootstrap"
  echo "Building $EXECUTABLE..."
  GOOS=linux GOARCH=amd64 go build \
    -tags lambda.norpc \
    -ldflags="-X 'main.Version=$2'" \
    -o "$OUTPUT_DIR/bootstrap" \
    "$CMD_DIR/$EXECUTABLE"

  # Check if the build was successful
  if [ $? -ne 0 ]; then
    echo "Failed to build $EXECUTABLE"
    exit 1
  fi

  # Zip the bootstrap into the dist directory
  echo "Zipping $EXECUTABLE..."
  (cd $OUTPUT_DIR && zip -r "../$EXECUTABLE.zip" bootstrap)
  rm -rf $OUTPUT_DIR

  # Check if the zip was successful
  if [ $? -ne 0 ]; then
    echo "Failed to zip $EXECUTABLE"
    exit 1
  fi

  echo "$EXECUTABLE has been successfully built and zipped."
}

# Array of executable names
executables=(
  "activate_agency"
  "create_agency"
  "deactivate_agency"
  "delete_agency"
  "list_agencies"
  "read_agency"
  "update_agency"
)

# Loop over each executable and call the build_and_zip_executable function
for EXECUTABLE in "${executables[@]}"; do
  build_and_zip_executable "$EXECUTABLE" "$SEMVER"
done

echo "All executables built and zipped successfully."
