# Sprinter
Sprinter is a web crawler that looks for links on the same domain.

## Installation
    make

Program accepts a flag `-root=<website>`.

## TODO
* Handle HTTP 429, 403, and other mechanisms related to rate-limiting.
* If we can find a sitemap, we should use it.
* Would be nice to add itests, and fuzz tests.