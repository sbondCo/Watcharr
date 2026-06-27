# Unreleased

These changes are awaiting release:

## Data migrations

This is the first time migrations are being taken place when updating Watcharr, please be mindful of that and **ensure you have your existing database backed up** incase of any errors.

### Backfilling `plays` data from users Activity.

This migration was added so that existing data could be used to fill out how many plays of your media you have. It can't be 100% accurate because it has to make some assumptions, but it can get close and saves you a lot of time.

### Dropping `deleted_at` columns for `watched_seasons` and `watched_episodes` tables since we do not use them.

I noticed that queries (eg for loading your main list) would take an extra ~100ms because we were adding a `deleted_at IS NULL` to them when getting watched seasons/episodes. We don't use the deleted_at column, so I have just removed it from the tables (and queries so we have the speed boost).

### Moving to using WAL journal_mode for our sqlite database, which will grant us improvements in all areas.

Most database operations will be much faster now and more concurrent use of the database is now possible.

#### Important Note About Backups

You will notice that the database now comprises of three files: `watcharr.db` (which has always been there) and `watcharr.db-wal`, `watcharr.db-shm` (which are new).

If you backup your database by copying the .db file (while your server is stopped of course), you should also copy the `watcharr.db-wal` file since it can contain database content. You can ignore the `watcharr.db-shm` file if you want.

## Added

- DB: Add custom migrations support.
- Activity: New `CountAsPlay` property for tracking media total plays (this will speed up counting plays since we'll no longer rely on text searches in the db).
  - Migration: Backfill media plays data from existing user activity, so no one has to start from `0` plays when they already have data we can use to get their total plays.
- Add link to names of top crew members.
- Return `Plays` in WatchedDto where used.
- Show `plays` count for media in `MyReview` component.
- Activity: Show icon for activity that counts as a play.

## Changed

- Moved db to WAL journal_mode.
- WatchedUpdateRequest: Manually validate instead of using complex struct tags.
  - Now properly validating WatchedStatus.

## Fixed

- fix safari: Shrikhand font.
- Icon: Fix status icons not having width and height properties not set.
- Status: Fix button sizes now that icons have a width/height set.
- Nav: Fix logo link being clickable through whole left side of nav.
- import: myanimelist: Don't import start/finish dates when they are empty.
- Star and Play icons color.
- Activity: Fixed automation tooltip going out of bounds by moving it to top.

## Removed

- Drop `deleted_at` columns for watched episode/season tables.
- Removed `AddActivity` (POST /activity) endpoint.
