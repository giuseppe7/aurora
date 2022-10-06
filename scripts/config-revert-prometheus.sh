#!/bin/bash

configFile="configs/prometheus/prometheus.yml"
if [ -f "$configFile.bak" ]; then
    mv configs/prometheus/prometheus.yml.bak configs/prometheus/prometheus.yml 
fi