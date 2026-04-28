#!/bin/bash
set -e

kafka-topics \
  --bootstrap-server "$KAFKA_BOOTSTRAP" \
  --create --if-not-exists \
  --topic "$KAFKA_TASK_TOPIC" \
  --partitions 1 \
  --replication-factor 1

kafka-topics \
  --bootstrap-server "$KAFKA_BOOTSTRAP" \
  --create --if-not-exists \
  --topic "$KAFKA_RESULT_TOPIC" \
  --partitions 1 \
  --replication-factor 1

echo "topics ready"
