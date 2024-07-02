---
sidebar_position: 100
description: Backing up your server.
---

# Backing Up

Watcharr cannot handle backing up for you, but its simple structure should make it easy to slide it into your existing backup routine.

Once you have created your first backup, it's recommended that you try restoring from that backup to make sure everything is correct and will work in case you need to restore from it in the future.

:::danger Shutdown your server!
Always shutdown Watcharr before backing up its data, otherwise you run the risk of corrupting files (mainly your database).
:::

## Simple Steps

:::success Recommended
This is the recommended way to backup.
:::

Here are the steps that are run daily for the `beta.watcharr.app` instance.

1. Shutdown Watcharr.
2. Copy and backup the entire `data` folder (default location is `./data`, backup whichever folder your server is configured to store its data in).
3. Start Watcharr.

I call these the simple steps because they simply backup your entire server, if you ever need to restore your server, it will start up from its exact state at backup.

## Advanced Steps

:::danger Discouraged
Backing up this way is discouraged because new important files could be added later and missed.
:::

If you don't care about the warning not to backup this way, here are the "important" files that you can single out for backup:

- `watcharr.db` Your database, holds all users, their watchlists, etc.
- `watcharr.json` Server config.
- `img/up` Profile picture uploads.
- `img/games` Game posters (not exactly important, but scenarios in which this folder is not backed up have not been tested, only relevant for servers with game support enabled).
