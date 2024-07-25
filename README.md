<h1 align="center">Watcharr</h1>
<p align="center"><img src="./static/logo-col.png" alt="logo" width="250" /></p>

<p align="center">
  <a href="https://github.com/sbondCo/Watcharr/pkgs/container/watcharr"><img src="https://img.shields.io/github/v/release/sbondCo/Watcharr?label=version&style=for-the-badge" /></a>
  <a href="https://beta.watcharr.app"><img src="https://img.shields.io/website?label=DEMO&style=for-the-badge&url=https%3A%2F%2Fbeta.watcharr.app" /></a>
  <a href="https://watcharr.app"><img src="https://img.shields.io/website?label=DOCS&style=for-the-badge&url=https%3A%2F%2Fwatcharr.app" /></a>
  <a href="https://github.com/sbondCo/Watcharr/issues"><img src="https://img.shields.io/github/issues-raw/sbondCo/Watcharr?label=ISSUES&style=for-the-badge" /></a>
  <a href="/LICENSE"><img src="https://img.shields.io/github/license/sbondCo/Watcharr?style=for-the-badge" /></a>
  <a href="https://matrix.to/#/#watcharr:matrix.org"><img src="https://img.shields.io/matrix/watcharr%3Amatrix.org?style=for-the-badge&logo=matrix" /></a>
</p>

I'm your new easily self-hosted content watched list. The place you store your watched (or watching, planned, etc) **movies** and **tv shows** (and **anime**), rate them and track their status.

With [some extra configuration](https://watcharr.app/docs/server_config/game-support-igdb) I can also track your **video games**.

I am built with Go and Svelte(Kit).

Feel free to abuse this demo instance (nicely), which runs on the latest `dev` build (there may be bugs, as new features are tested on here too): [https://beta.watcharr.app/](https://beta.watcharr.app/)

[Track Progress Until Next Version](https://github.com/orgs/sbondCo/projects/9/views/3)

### Contents

- [Screenshots](#screenshots)
- [Setup](#set-up)
- [Contributing](CONTRIBUTING.md)
- [Community Made Tools](#community-made-tools)
- [Getting Help](#getting-help)

# Screenshots

<p align="center">

| Homepage                                                   | Watched Show Hover                                                      |
| ---------------------------------------------------------- | ----------------------------------------------------------------------- |
| <img src="./screenshot/homepage.png" alt="Watched List" /> | <img src="./screenshot/homepage-poster-hover.png" alt="Watched List" /> |

| Watched Show Status Change                                                              | Movie Details                                                                  |
| --------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------ |
| <img src="./screenshot/homepage-poster-change-status.png" alt="Changing Show Status" /> | <img src="./screenshot/content-details-page.png" alt="Content Details Page" /> |

| User Profile                                                        | Discover                                                         |
| ------------------------------------------------------------------- | ---------------------------------------------------------------- |
| <img src="./screenshot/user-profile.png" alt="User Profile Page" /> | <img src="./screenshot/discover-page.png" alt="Discover Page" /> |

| Dark Homepage                                                          | Dark Content Details                                                                   |
| ---------------------------------------------------------------------- | -------------------------------------------------------------------------------------- |
| <img src="./screenshot/homepage-dark.png" alt="Dark Theme Homepage" /> | <img src="./screenshot/content-details-page-dark.png" alt="Dark Theme Content Page" /> |

</p>

# Set Up

[Checkout our documentation](https://watcharr.app/docs/category/installation) for an up to date guide on setup! If you hate manuals, but love docker, this [docker-compose.yml](./docker-compose.yml) file is your friend.

# Community Made Tools

Third-party tools made by the community for enhancing your Watcharr experience!

- [Kodi Plugin](https://github.com/airdogvan/watcharr_kodi) by [airdogvan](https://github.com/airdogvan) for automatically tracking your watched shows/movies.

Thanks to anyone that has made a script or tool for Watcharr. Feel free to add your own to the list if you have one!

**Note:** I cannot provide any assurances for these tools or stay on top of them (code review, etc), if you have any problems please open an issue in the project for the tool so that they can stay organized.

# Getting Help

If something isn't working for you or you are stuck, [creating an issue](https://github.com/sbondCo/Watcharr/issues/new) is the best way to get help! Every type of issue is accepted, so don't be afraid to ask anything!

You can also [join our space on Matrix](https://matrix.to/#/#watcharr:matrix.org) for support.
