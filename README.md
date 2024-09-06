# image-uploader

To run create a docker-compose.yml file with the following content:

```docker-compose
services:
  image-uploader:
    image: savarez/image-uploader:latest
    ports:
      - "8086:8086"
    volumes:
      - ./uploads:/app/uploads
    container_name: image-uploader
```

and folder 'uploads' must exist.

```bash
mkdir uploads
```

First run:

```bash
docker-compose up --build
```
On first run TOTP secret will be generated and shown

If everything is ok, the container can be started in background:

```bash
docker-compose up -d
```