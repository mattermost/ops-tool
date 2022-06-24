#!/bin/bash

timeout --preserve-status 1 \
curl -s 'https://v2.jokeapi.dev/joke/Coding?type=single&format=json' \
| jq --slurp -c '{data: .}'
