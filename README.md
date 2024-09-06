# image-uploader

To run create a docker-compose.yml file with the following content:

```docker-compose
services:
  image-uploader:
    image: savarez/image-uploader:latest
    ports:
      - "8086:8086"
    environment:
        - IMAGES_URL=https://yourhostname/images
    volumes:
      - ./uploads:/app/uploads
      - ./totp_secret:/app/totp_secret
    container_name: image-uploader
```


First run:

```bash
docker pull savarez/image-uploader:latest
docker-compose up
```
On first run TOTP secret will be generated and shown

If everything is ok, the container can be started in background:

```bash
docker-compose up -d
```