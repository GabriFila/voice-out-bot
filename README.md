# Voice Out Bot

This is a Telegram Bot to allow users to share telegram voice notes outside telegram, e.g. Whatsapp.

To achieve it, you simply share the voice note with this bot and it will reply you back with a shareable audio file to share out the app.

## Privacy

The bot doesn't store any info or data about the user or the sent messages.

## Development

The bot is built in [Go](https://go.dev/) using the [Telebot](https://github.com/tucnak/telebot) framework.

## Deployment

The bot is deployed on Google Cloud Run and it follows the [buildpack](https://buildpacks.io/) convention to build the Go container for deployment. The Dockerfile is there but it is not used during the deploy pipeline.
