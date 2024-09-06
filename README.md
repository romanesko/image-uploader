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
      - ./secrets:/app/secrets
    container_name: image-uploader
```


Run:

```bash
docker-compose up -d
```

file with TOTP secret should be created in secrets folder.