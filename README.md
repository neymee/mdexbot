## About
A telegram bot that can notify you about new manga chapters on [MangaDex](https://mangadex.org).

![image](https://github.com/neymee/mdexbot/actions/workflows/go.yml/badge.svg)

## How to run
1. Go to the [BotFather](https://t.me/BotFather) and create a new bot.
2. Copy `config.json` from `./config` directory to project root. Fill in the `bot.token` field with your bot's token.
3. Run the infrastructure:
```bash
cd deployments
docker compose run -d
```
4. Run the bot:
```bash
# from the project root
go run ./cmd/app/main.go 
```
Optionally, you can specify the following environment variables: 
- `PRETTY_LOGGING=true` to make logs more human readable;
- `CONFIG_PATH=path/to/config.json` to specify the path to the config file.
