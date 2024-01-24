# ICS adapters

This web service provides iCalendar (RFC 5545) feeds for certain websites that
don't have such a feed already. These can be subscribed by any calendar app,
such as

- [ICSx⁵](https://icsx5.bitfire.at/) for Android
- [gnome-calendar](https://wiki.gnome.org/Apps/Calendar)
- Google Calendar

## List of feeds

The full list is auto-generated on the index page: https://ics.cbix.de/

## Running the web service

This is the default mode when running the build:

```
go build
./ics
```

## Static site generator

Not supported yet.
