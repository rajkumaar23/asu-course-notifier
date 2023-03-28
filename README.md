# asu-course-notifier

To run asu-course-notifier on your system, follow the steps below :

- Ensure you have `docker` and `docker-compose` installed.
- Create a directory for storing the config and docker-compose files.
- Get inside *that* directory.
- Create `config.json` with the contents resembling the format of this [example config file](https://github.com/rajkumaar23/asu-course-notifier/blob/main/config.example.json).
- Create `docker-compose.yml` inside the same directory with the content below :
  - **DO NOT FORGET** to update the `/absolute/path/to/config.json` inside the compose file
```yaml
version: "3"

services:
  asu-course-notifier:
    image: ghcr.io/rajkumaar23/asu-course-notifier:main
    container_name: asu-course-notifier
    restart: unless-stopped
    volumes:
      - /absolute/path/to/config.json:/app/config.json
      - /etc/timezone:/etc/timezone:ro
      - /etc/localtime:/etc/localtime:ro
```
- Spin up the container with `docker-compose up -d` and enjoy!
