#! /usr/bin/env bash

set -e

BASE_DIR="sources"
TARGETS=()

while [[ $# -gt 0 ]]; do
  KEY="$1"
  case $KEY in
    --base-dir)
        BASE_DIR=$2
        shift
        shift
        ;;
    *)
        TARGETS+=($KEY)
        shift
        ;;
  esac
done

if [ ${#TARGETS[@]} -eq 0 ]; then
  echo "No targets specified; building all."
  TARGETS=( $(ls $BASE_DIR) )
fi

echo "Have ${#TARGETS[@]} targets."

cd $BASE_DIR

for TARGET in ${TARGETS[@]}; do
  echo "At $TARGET."

  if [ -z "$BASE_DIR/$TARGET/diagram.py" ]; then
    continue
  fi

  cd $TARGET
  python3 diagram.py
  cd $BASE_DIR
done