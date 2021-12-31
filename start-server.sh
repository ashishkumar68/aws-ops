#!/bin/bash

if [[ $AWS_ACCESS_KEY_ID == "" ]]; then
  printf "Could not find AWS_ACCESS_KEY_ID\n"
  exit 1
fi

if [[ $AWS_SECRET_ACCESS_KEY == "" ]]; then
  printf "Could not find AWS_SECRET_ACCESS_KEY\n"
  exit 1
fi

if [[ $AWS_DEFAULT_REGION == "" ]]; then
  printf "Could not find AWS_DEFAULT_REGION\n"
  exit 1
fi

printf "Starting containers..\n"

docker-compose up -d --build --force-recreate

sleep 10

printf "Containers are ready..\n"