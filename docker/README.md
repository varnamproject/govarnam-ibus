# Docker

Builds `govarnam` & `govarnam-ibus` for Ubuntu 14.04 (GTK 3.10, glibc 2.19) & later versions:
```bash
docker build -t varnam .
```

Copy built artifacts from container using:
```bash
id=$(docker create varnam)
docker cp $id:/extract/. ./
docker rm -v $id
```

Docker's new version has `buildx`. With it you can cross-compile:
```bash
docker buildx build -t varnam --platform=linux/i386 .
```