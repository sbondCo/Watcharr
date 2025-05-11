---
sidebar_position: 1
---

# Text File (.txt list)

:::info Backup?
You may consider a backup of your server before starting any import. They are not easily reversible, though we do our best to ensure they are accurate and bug free!
:::

The text file (.txt) import is of an arbitrary format (the one I used for years before creating Watcharr).

Hopefully it is useful for others with similar files or in scenarios where its the easiest to generate for an import (though if possible, when generating a backup from another service data manually, matching a Watcharr export would enable keeping more data).

## Format

Each line is a new entry. The name of the content (show/movie) must be provided, the year can be optionally provided surrounded by brackets. Doesn't support specifying if name is for a show or movie, the importer will only automatically match on full search matches, if there are multiple results, you will be asked to pick the correct one.

```
<name> [(<year>)]
```

## An example

```
The Terminator (1984)
Breaking Bad
A Fistful of Dollars
Reacher (2022)
```
