# workVerification

Go service that consumes APK work verification requests from Kafka, downloads files from MinIO, scans them with ClamAV and YARA, and publishes the result back to Kafka.

## Environment variables

- `MINIO_ENDPOINT` default `localhost:9000`
- `MINIO_ACCESS_KEY` default `admin`
- `MINIO_SECRET_KEY` default `password123`
- `MINIO_BUCKET` default `app-builds`
- `KAFKA_BROKER` default `localhost:9092`
- `CLAMD_HOST` default `localhost`
- `CLAMD_PORT` default `3310`

## Run

```bash
go run ./cmd/app
```
