# fxgo

> [!WARNING]
> Heavy WIP, not ready for use. If you're interested, check again in a few weeks or months.

> [!IMPORTANT]
> Please note that this API is not intended to provide real-time data. Updates are constrained by the sources, which typically publish batches once per day.
> The expected use case is something like a finance tracker, where you want to add a transaction in a foreign currency and need it roughly translated into your base currency.

## Motivation

I've been developing a financial tracker for my own use and ran into the question of getting reliable-ish FX data for it. I don't want to use a public API, and was looking for a self-hosted solution that clearly states its sources and could be easily maintained by me if the author stops working on it. Active maintenance is important, as source formats may change. See [relevant projects](#relevant-projects) for what I've considered.

## Maintenance

A project like this will most likely require regular maintenance. I'll maintain it while I continue using my financial tracker and don't plan on stopping anytime soon.

I aim to write the project in a way that makes it easy for anyone familiar with Go to pick up and maintain.

## Sources and Reliability

I plan to use sources such as official central banks and federal reserves. Institutions usually post data once per day, and each currency has its preferred source.

## Relevant Projects

[frankfurter](https://github.com/lineofflight/frankfurter) - amazing project that provides data from the European Central Bank in a nice way, and what I was using before. However, it has a very small number of currencies available, and is written in Ruby, in which I have no experience.

[exchange-api](https://github.com/fawazahmed0/exchange-api) - 200+ currencies, but the sources are unclear. Most likely uses scrapers on schedule. Not exactly self-hostable as the author doesn't include the scraper setup, but might be good enough if you just want the data for a small-scale project.

## Contribution

PRs and issues are highly welcome.
