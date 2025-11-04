![enbyware](https://pride-badges.pony.workers.dev/static/v1?label=enbyware&labelColor=%23555&stripeWidth=8&stripeColors=FCF434%2CFFFFFF%2C9C59D1%2C2C2C2C)

**tlynxmap** renders journeys exported from Travelynx to a PNG file.

Optionally, it can filter by time and exclude certain stations. A rendered image might look like this:

![A map centered on Germany showing various journeys between Passau, Frankfurt, Karlsruhe, Bonn, Aachen, and some reaching into the Netherlands in Utrecht, Amsterdam, and Almere. The bottom is captioned "Maps and Data (c) openstreetmap.org and contributors, ODbL, rendered using tlynxmap". The lower right corner features an enby flag.](./example.png)

Or, in dark mode:

![The same map as above, but in dark mode. The attribution string is different: "Map (c) Carto [CC BY 3.0] Data (c) OSM and contributors, ODbL., rendered using tlynxmap".](./example-dark.png)

## Building

```sh
go build
```

## Usage

```
Usage of ./tlynxmap:
  -dark
    	Render using a dark theme
  -data string
    	Travelynx raw data file (default "data.json")
  -end string
    	Include journeys until this date (optional)
  -exclude string
    	File with station name substrings to exclude (optional)
  -height uint
    	Height of the generated image (default 1000)
  -hide-attribution
    	Hide the attribution string (warning: may violate licenses if published)
  -output string
    	Output file for the renderend PNG (default "output.png")
  -start string
    	Include journeys from this date (optional)
  -width uint
    	Width of the generated image (default 1800)
```
