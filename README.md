# asu-course-notifier

## Docker

To run asu-course-notifier on any system, pull the built image from GitHub container registry:


```bash
docker pull ghcr.io/rajkumaar23/asu-course-notifier:main
```

Ensure to create `config.json` by copying the format from `config.example.json`. Update the necessary values & pass it when running the container:
```bash
docker run -d -v /absolute/path/to/config.json:/app/config.json asu-course-notifier
```
