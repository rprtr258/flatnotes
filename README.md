<p align="center">
  <img src="docs/logo.svg" width="300px"></img>
</p>
<p align="center">
  <img alt="Docker Pulls" src="https://img.shields.io/docker/pulls/dullage/flatnotes?style=for-the-badge">
</p>

A self-hosted, database-less note-taking web app that utilises a flat folder of markdown files for storage.

Log into the [demo site](https://demo.flatnotes.io) and take a look around. *Note: This site resets every 15 minutes.*

## Contents

* [Design Principle](#design-principle)
* [Features](#features)
* [Getting Started](#getting-started)
  * [Using Docker](#using-docker)
  * [Using Docker Compose](#using-docker-compose)

## Design Principle
flatnotes is designed to be a distraction-free note-taking app that puts your note content first. This means:

* A clean and simple user interface.
* No folders, notebooks or anything like that. Just all of your notes, backed by powerful search and tagging functionality.
* Quick access to a full-text search from anywhere in the app (keyboard shortcut "/").

Another key design principle is not to take your notes hostage. Your notes are just markdown files. There's no database, proprietary formatting, complicated folder structures or anything like that. You're free at any point to just move the files elsewhere and use another app.

Equally, the only thing flatnotes caches is the search index and that's incrementally synced on every search (and when flatnotes first starts). This means that you're free to add, edit & delete the markdown files outside of flatnotes even whilst flatnotes is running.

## Features
* Advanced search functionality.
* Note "tagging" functionality.
* Light/dark themes.
* Restful API.

See [the wiki](https://github.com/rprtr258/flatnotes/wiki) for more details.

## Getting Started
### Using Docker
```shell
docker run -d \
  -e "PUID=1000" \
  -e "PGID=1000" \
  -e "FLATNOTES_AUTH_TYPE=password" \
  -e "FLATNOTES_USERNAME=user" \
  -e "FLATNOTES_PASSWORD=changeMe!" \
  -e "FLATNOTES_SECRET_KEY=aLongRandomSeriesOfCharacters" \
  -v "$(pwd)/data:/data" \
  -p "8080:8080" \
  dullage/flatnotes:latest
```

### Using Docker Compose
[docker-compose.yaml](./docker-compose.yaml)

See the [Environment Variables](https://github.com/rprtr258/flatnotes/wiki/Environment-Variables) article in the wiki for a full list of configuration options.

