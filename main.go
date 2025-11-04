package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"image/color"
	"os"
	"strconv"
	"strings"
	"time"

	sm "github.com/flopp/go-staticmaps"
	"github.com/fogleman/gg"
	"github.com/golang/geo/s2"
)

var enbyFlag = []color.Color{
	color.NRGBA{R: 0xff, G: 0xf4, B: 0x2f, A: 0xff},
	color.White,
	color.NRGBA{R: 0x9c, G: 0x59, B: 0xd1, A: 0xff},
	color.NRGBA{R: 0x29, G: 0x29, B: 0x29, A: 0xff},
}

type Theme interface {
	CreateTileProvider() *sm.TileProvider
	GetColorExact() color.Color
	GetColorBeeline() color.Color
	GetColorStation() color.Color
}

type Options struct {
	Width           int
	Height          int
	Dark            bool
	HideAttribution bool

	StartDate        time.Time
	EndDate          time.Time
	ExcludedStations []string

	DataFile   string
	OutputFile string
}

type Data struct {
	Journeys []*Journey
}

type Journey struct {
	RealDepTS string `json:"real_dep_ts"`
	RealArrTS string `json:"real_arr_ts"`

	RealDepTSNormalized time.Time
	RealArrTSNormalized time.Time

	DepName string  `json:"dep_name"`
	DepLat  float64 `json:"dep_lat"`
	DepLon  float64 `json:"dep_lon"`
	DepEva  int     `json:"dep_eva"`

	ArrName string  `json:"arr_name"`
	ArrLat  float64 `json:"arr_lat"`
	ArrLon  float64 `json:"arr_lon"`
	ArrEva  int     `json:"arr_eva"`

	Polyline string `json:"polyline"`
	Route    string `json:"route"`

	PolylineNormalized []PolylinePoint

	Hidden bool
}

func (j *Journey) Exact() bool {
	return j.Polyline != ""
}

type Route [][]any

type Polyline [][]any

type PolylinePoint struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
	Eva int     `json:"eva"`
}

func parseDateFlag(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}

	return time.Parse(time.DateOnly, s)
}

func readExcludeFile(path string) ([]string, error) {
	if path == "" {
		return []string{}, nil
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var stations []string
	for line := range strings.Lines(string(b)) {
		line = strings.TrimSpace(line)
		if line != "" {
			stations = append(stations, line)
		}
	}

	return stations, nil
}

func parseCommandLine() (Options, error) {
	var flagWidth uint
	var flagHeight uint
	var flagDark bool
	var flagHideAttribution bool

	var flagStartDate string
	var flagEndDate string
	var flagExcludeFile string

	var flagDataFile string
	var flagOutputFile string

	flag.UintVar(&flagWidth, "width", 1800, "Width of the generated image")
	flag.UintVar(&flagHeight, "height", 1000, "Height of the generated image")
	flag.BoolVar(&flagDark, "dark", false, "Render using a dark theme")
	flag.BoolVar(&flagHideAttribution, "hide-attribution", false, "Hide the attribution string (warning: may violate licenses if published)")

	flag.StringVar(&flagStartDate, "start", "", "Include journeys from this date (optional)")
	flag.StringVar(&flagEndDate, "end", "", "Include journeys until this date (optional)")
	flag.StringVar(&flagExcludeFile, "exclude", "", "File with station name substrings to exclude (optional)")

	flag.StringVar(&flagDataFile, "data", "data.json", "Travelynx raw data file")
	flag.StringVar(&flagOutputFile, "output", "output.png", "Output file for the renderend PNG")

	flag.Parse()

	startDate, err := parseDateFlag(flagStartDate)
	if err != nil {
		return Options{}, fmt.Errorf("error parsing start date: %w", err)
	}

	endDate, err := parseDateFlag(flagEndDate)
	if err != nil {
		return Options{}, fmt.Errorf("error parsing end date: %w", err)
	}

	excludedStations, err := readExcludeFile(flagExcludeFile)
	if err != nil {
		return Options{}, fmt.Errorf("error reading stations exclude file: %w", err)
	}

	return Options{
		Width:           int(flagWidth),
		Height:          int(flagHeight),
		Dark:            flagDark,
		HideAttribution: flagHideAttribution,

		StartDate:        startDate,
		EndDate:          endDate,
		ExcludedStations: excludedStations,

		DataFile:   flagDataFile,
		OutputFile: flagOutputFile,
	}, nil
}

func processTlynxData(options Options, data *Data) error {
JOURNEYS:
	for _, journey := range data.Journeys {
		for _, s := range options.ExcludedStations {
			if strings.Contains(journey.ArrName, s) || strings.Contains(journey.DepName, s) {
				journey.Hidden = true
				continue JOURNEYS
			}
		}

		t, err := strconv.ParseFloat(journey.RealDepTS, 64)
		if err != nil {
			return fmt.Errorf("journey from %s to %s: error parsing real dep. time: %w", journey.DepName, journey.ArrName, err)
		}

		journey.RealDepTSNormalized = time.Unix(int64(t), 0)

		t, err = strconv.ParseFloat(journey.RealArrTS, 64)
		if err != nil {
			return fmt.Errorf("journey from %s to %s: error parsing real arr. time: %w", journey.DepName, journey.ArrName, err)
		}

		journey.RealArrTSNormalized = time.Unix(int64(t), 0)

		if !options.StartDate.IsZero() && journey.RealArrTSNormalized.Before(options.StartDate) {
			journey.Hidden = true
		}

		if !options.EndDate.IsZero() && journey.RealDepTSNormalized.After(options.EndDate) {
			journey.Hidden = true
		}

		if journey.Polyline == "" {
			var route Route
			if err := json.Unmarshal([]byte(journey.Route), &route); err != nil {
				return fmt.Errorf("journey from %s to %s: error parsing route: %w", journey.DepName, journey.ArrName, err)
			}

			for _, routePoint := range route {
				details := routePoint[2].(map[string]any)

				polylinePoint := PolylinePoint{
					Lat: details["lat"].(float64),
					Lon: details["lon"].(float64),
				}

				eva := routePoint[1]
				switch eva := eva.(type) {
				case string:
					polylinePoint.Eva, err = strconv.Atoi(eva)
					if err != nil {
						return fmt.Errorf("journey from %s to %s: error parsing string eva in route: %w", journey.DepName, journey.ArrName, err)
					}
				case float64:
					polylinePoint.Eva = int(eva)
				default:
					return fmt.Errorf("journey from %s to %s: eva in route has unexpected type: %#v", journey.DepName, journey.ArrName, eva)
				}

				journey.PolylineNormalized = append(journey.PolylineNormalized, polylinePoint)
			}
		} else {
			var polyline Polyline
			if err := json.Unmarshal([]byte(journey.Polyline), &polyline); err != nil {
				return fmt.Errorf("journey from %s to %s: error parsing polyline: %w", journey.DepName, journey.ArrName, err)
			}

			for _, polylinePointRaw := range polyline {
				polylinePoint := PolylinePoint{
					Lat: polylinePointRaw[1].(float64),
					Lon: polylinePointRaw[0].(float64),
				}

				if len(polylinePointRaw) > 2 {
					eva := polylinePointRaw[2]
					switch eva := eva.(type) {
					case string:
						polylinePoint.Eva, err = strconv.Atoi(eva)
						if err != nil {
							return fmt.Errorf("journey from %s to %s: error parsing string eva in polyline: %w", journey.DepName, journey.ArrName, err)
						}
					case float64:
						polylinePoint.Eva = int(eva)
					default:
						return fmt.Errorf("journey from %s to %s: eva in polyline has unexpected type: %#v", journey.DepName, journey.ArrName, eva)
					}
				}

				journey.PolylineNormalized = append(journey.PolylineNormalized, polylinePoint)
			}
		}
	}

	return nil
}

func renderMap(options Options, data Data) error {
	ctx := sm.NewContext()
	ctx.SetSize(options.Width, options.Height)

	var theme Theme

	if options.Dark {
		theme = &ThemeDark{}
	} else {
		theme = &ThemeLight{}
	}

	ctx.SetTileProvider(theme.CreateTileProvider())

	if options.HideAttribution {
		ctx.OverrideAttribution("")
	} else {
		ctx.OverrideAttribution(ctx.Attribution() + ", rendered using tlynxmap")
	}

	for _, journey := range data.Journeys {
		if journey.Hidden {
			continue
		}

		col := theme.GetColorBeeline()
		if journey.Exact() {
			col = theme.GetColorExact()
		}

		lls := make([]s2.LatLng, 0, len(journey.PolylineNormalized))
		boarded := false
		for _, point := range journey.PolylineNormalized {
			if boarded || point.Eva == journey.DepEva {
				boarded = true
				ll := s2.LatLngFromDegrees(point.Lat, point.Lon)
				lls = append(lls, ll)
				if point.Eva == journey.ArrEva {
					break
				}
			}
		}

		path := sm.NewPath(lls, col, 2)
		ctx.AddObject(path)
	}

	for _, journey := range data.Journeys {
		if journey.Hidden {
			continue
		}

		addStationCircle(theme, ctx, journey.DepLat, journey.DepLon)
		addStationCircle(theme, ctx, journey.ArrLat, journey.ArrLon)
	}

	img, err := ctx.Render()
	if err != nil {
		return fmt.Errorf("error rendering map: %w", err)
	}

	gctx := gg.NewContextForImage(img)

	if !options.HideAttribution {
		addEnbyFlag(options, gctx)
	}

	if err := gctx.SavePNG(options.OutputFile); err != nil {
		return fmt.Errorf("error saving PNG: %w", err)
	}

	return nil
}

func addEnbyFlag(options Options, gctx *gg.Context) {
	cw := float64(gctx.Width())
	ch := float64(gctx.Height())

	boxHeight := gctx.FontHeight() + 4.
	width := 20.
	stripeHeight := boxHeight / 4.

	for i, color := range enbyFlag {
		gctx.SetColor(color)
		gctx.DrawRectangle(cw-width, ch-boxHeight+stripeHeight*float64(i), width, stripeHeight)
		gctx.Fill()
	}

	// Dim the flag a bit
	dimAlpha := 0x05
	if options.Dark {
		dimAlpha = 0x20
	}
	gctx.SetColor(color.NRGBA{A: uint8(dimAlpha)})
	gctx.DrawRectangle(cw-width, ch-boxHeight, width, boxHeight)
	gctx.Fill()
}

func addStationCircle(theme Theme, ctx *sm.Context, lat, lon float64) {
	ll := s2.LatLngFromDegrees(lat, lon)
	ctx.AddObject(sm.NewCircle(ll, color.Transparent, theme.GetColorStation(), 3000, 1))
}

func run() error {
	options, err := parseCommandLine()
	if err != nil {
		return err
	}

	dataBytes, err := os.ReadFile(options.DataFile)
	if err != nil {
		return fmt.Errorf("error reading data file: %w", err)
	}

	var data Data
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return fmt.Errorf("error parsing data file: %w", err)
	}

	if err := processTlynxData(options, &data); err != nil {
		return err
	}

	return renderMap(options, data)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}
}
