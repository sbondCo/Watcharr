# End-to-end (e2e) tests for web

Browser tests written in plain English for [Kane CLI](https://www.testmuai.com/kane-cli/).

| Test | What it covers |
|---|---|
| `setup_test.md` | Creating the admin user, checking the profile menu, logging out and back in |
| `watched_test.md` | Admin setup, searching for a movie, marking it as finished and seeing it in the list |

## Run

Start a fresh instance on port 4840:

```bash
docker run -d --name watcharr-e2e -p 4840:3080 ghcr.io/sbondco/watcharr:latest
npm install -g @testmuai/kane-cli
kane-cli login
kane-cli testmd run e2e/kane/setup_test.md
kane-cli testmd run e2e/kane/watched_test.md
```

Both tests start from the admin setup page, so recreate the container before each run. `watched_test.md` needs internet access for movie search.
