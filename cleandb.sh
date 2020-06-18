#!/usr/bin/env bash

read -p "Cleaning Prometheus as well as Grafana Databases... Continue? (Y/N): " confirm

if [[ $confirm != [Yy] ]]; then
    echo "Input was [Nn] or not recognizable... stopping."
    exit 1
fi

echo "Cleaning databases..."

rm -rf ./data/grafana/*
rm -rf ./data/prometheus/*
rm -rf ./data/thresholds/*

if [[ $? -ne 0 ]]; then
    echo "Error occured... exiting ."
    exit 1
fi

echo "Done... exiting."
exit 0
