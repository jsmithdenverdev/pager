#!/bin/bash

# Set the directory where the executables are located
CMD_DIR="./cmd"
DIST_DIR="./dist"

# Create the dist directory if it doesn't exist
mkdir -p $DIST_DIR

# Function to build and zip an executable
build_and_zip_executable() {
  EXECUTABLE=$1

  # Create the output directory for the executable
  OUTPUT_DIR="$DIST_DIR/$EXECUTABLE"
  mkdir -p $OUTPUT_DIR

  # Build the executable and output to the dist directory as "handler"
  echo "Building $EXECUTABLE..."
  go build -o "$OUTPUT_DIR/handler" "$CMD_DIR/$EXECUTABLE"

  # Check if the build was successful
  if [ $? -ne 0 ]; then
    echo "Failed to build $EXECUTABLE"
    exit 1
  fi

  # Zip the handler into the dist directory
  echo "Zipping $EXECUTABLE..."
  (cd $OUTPUT_DIR && zip -r "../$EXECUTABLE.zip" handler)

  # Check if the zip was successful
  if [ $? -ne 0 ]; then
    echo "Failed to zip $EXECUTABLE"
    exit 1
  fi

  echo "$EXECUTABLE has been successfully built and zipped."
}

# Array of executable names
executables=("create_agency")

# Loop over each executable and call the build_and_zip_executable function
for EXECUTABLE in "${executables[@]}"; do
  build_and_zip_executable "$EXECUTABLE"
done

echo "All executables built and zipped successfully."
