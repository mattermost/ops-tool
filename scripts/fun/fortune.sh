#!/bin/bash

fortune | tr -dc '[[:print:]]' | jq -R -c '{"data": [{"fortune": .}]}'
