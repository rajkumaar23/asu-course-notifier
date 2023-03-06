# asu-course-notifier

## Docker

To run asu-course-notifier on any system, pull the built image from GitHub container registry:


```docker
docker pull ghcr.io/rajkumaar23/asu-course-notifier:main
```

Ensure to create `config.json` by copying the format from `config.example.json`. Update the necessary values inside config.json & pass it as a volume when running the container:
```sh
docker run -d \
  -v /absolute/path/to/config.json:/app/config.json \
  -v /etc/timezone:/etc/timezone:ro -v /etc/localtime:/etc/localtime:ro \
  ghcr.io/rajkumaar23/asu-course-notifier:main
```
