package main

import (
	"image/color"

	sm "github.com/flopp/go-staticmaps"
)

type ThemeLight struct{}

func (t *ThemeLight) CreateTileProvider() *sm.TileProvider {
	return sm.NewTileProviderOpenStreetMaps()
}

func (t *ThemeLight) GetColorBeeline() color.Color {
	return color.NRGBA{
		R: 0x66,
		G: 0x55,
		B: 0x77,
		A: .6 * 0xff,
	}
}

func (t *ThemeLight) GetColorExact() color.Color {
	return color.NRGBA{
		R: 0x67,
		G: 0x3a,
		B: 0xb7,
		A: .8 * 0xff,
	}
}

func (t *ThemeLight) GetColorStation() color.Color {
	return color.NRGBA{
		R: 0xff,
		G: 0x00,
		B: 0x33,
		A: .4 * 0xff,
	}
}

var _ Theme = (*ThemeLight)(nil)

type ThemeDark struct{}

func (t *ThemeDark) CreateTileProvider() *sm.TileProvider {
	return sm.NewTileProviderCartoDark()
}

func (t *ThemeDark) GetColorBeeline() color.Color {
	return color.NRGBA{
		R: 0x66,
		G: 0x55,
		B: 0x77,
		A: .6 * 0xff,
	}
}

func (t *ThemeDark) GetColorExact() color.Color {
	return color.NRGBA{
		R: 0x58,
		G: 0x30,
		B: 0x9f,
		A: .8 * 0xff,
	}
}

func (t *ThemeDark) GetColorStation() color.Color {
	return color.NRGBA{
		R: 0xaa,
		G: 0x00,
		B: 0x22,
		A: .4 * 0xff,
	}
}

var _ Theme = (*ThemeDark)(nil)
