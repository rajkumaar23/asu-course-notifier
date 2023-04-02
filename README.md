# asu-course-notifier

# Setup
To run asu-course-notifier on your system, follow the steps below :

- Ensure you have `docker` and `docker-compose` installed.
- Create a directory for storing the config and docker-compose files.
- Get inside *that* directory.
- Create `config.json` with the contents resembling the format of this [example config file](https://github.com/rajkumaar23/asu-course-notifier/blob/main/config.example.json). Refer the [configuration](#configuration) section for more details.
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
```
- Spin up the container with `docker-compose up -d` and enjoy!

# Configuration
```json
{
  "COURSES" : {
    "12345" : ["A", "B"]
  },
  "TELEGRAM_IDS" : {
    "A" : 123,
    "B" : 234
  },
  "TERM_ID": "1234",
  "DEPT_CODE": "CSE",
  "BOT_TOKEN": ""
}
```
Breaking down each of the options :
- `COURSES`
  - `12345` is the class number that you find on the course catalog. This is unique for each class (course, instructor, schedule, term).
  - `["A", "B"]` are the names of users who have to be notified when a slot opens up for the class : `12345`. Remember that, these names should have their corresponding telegram IDs mapped inside the `TELEGRAM_IDS` object.
- `TELEGRAM_IDS`
  - `"A" : 123` - Here, `"A"` can be any sorta name you would like to call that user. This name will be used to greet them when they receive an alert from the bot. And `123` is their Telegram user ID (**an integer**) which can only be found using another Telegram bot. For instance, one can use [Chat ID Bot](https://t.me/chat_id_echo_bot).

- `DEPT_CODE` & `TERM_ID`
  - When you visit ASU Class Search website and choose your semester, your address bar would look something like `https://catalog.apps.asu.edu/catalog/classes/classlist?campusOrOnlineSelection=C&honors=F&level=grad&promod=F&searchType=all&subject=CSE&term=2231`
    - `DEPT_CODE` in our configuration is same as the `subject` parameter (*CSE*) in this URL.
    - `TERM_ID` in our configuration is same as the `term` parameter (*2231*) in this URL.

- `BOT_TOKEN`
  - You will get the `BOT_TOKEN` when you create a Telegram bot. This is used to authenticate when sending the alerts. More about bots [here](https://core.telegram.org/bots).
