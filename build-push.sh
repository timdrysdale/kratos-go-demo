#!/bin/bash
docker build --tag pingbar .                             
docker image tag pingbar:latest practable/core-prac-io:pingbar-0.0
docker push practable/core-prac-io:pingbar-0.0
