# MavLink Demo (Proxylity)


This project aims to demonstrate the power of proxylity as an efficient UDP Gateway system, using MAVLINK as an example.

MAVLINK is a UDP protocol, used widely by IOT devices to stream telemetry data with metrics like Position, Health, Temperature, e.t.c.

This project demos a simpler example; we simulate drone location data, push to proxylity's gateway which in turn pushes to a `FIFO` queue on AWS.

We read from `FIFO` queue and write to active websocket connections, these websocket connections are initiated on the frontend application which renders
a map for showing device locations. We receive websocket update about a new device and update it's position live on the map.


## Architecture Diagram

![alt text](image.png)



## Overview

There are 3 components to making this demo work as expected:

- `mavlink-backend`: This repo contains logic for polling queue for data and streaming to websocket connection
- `mavlink-frontend`: Here, we have the map page which displays live locations of registered devices.
- `mavlink-simulator`: This tool when called, simulates `n` devices and streams their locations as mavlink packets to **proxylity's udp gateway**.


In this repo, we have the necessary scripts and templates to setup resources our components depend on.


## Deploying

> **NOTE**: The instructions below assume the `aws` CLI, `jq` and `sam` are available on your Linux system.\
> **Go** Version Requirement for this example is `1.25.4`.

To deploy the template:

```bash
sam deploy \
    --stack-name mavlink-device-stack \
    --template-file ./template.json \
    --capabilities CAPABILITY_IAM \
    --region us-west-2
```

Once deployed, to get the ouputs from the stack and store the salient values in environment variables:

```bash
aws cloudformation describe-stacks \
  --stack-name mavlink-device-stack \
  --query "Stacks[0].Outputs" \
  --region us-west-2 \
  > outputs.json 

export DOMAIN=$(jq -r ".[]|select(.OutputKey==\"Domain\")|.OutputValue" outputs.json)
export PORT=$(jq -r ".[]|select(.OutputKey==\"Port\")|.OutputValue" outputs.json)
export SQS_URL=$(jq -r ".[]|select(.OutputKey==\"QueueURL\")|.OutputValue" outputs.json)
```

---

Next, we build the remaining components using `docker compose ...`:


