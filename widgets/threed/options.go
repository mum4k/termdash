// Copyright 2026 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// threed/options.go

package threed

// Options represents configuration options for the ThreeD widget.
type Options struct {
	RotationStep    float64 // Rotation step size in radians
	ZoomScale       float64 // Initial zoom scale for the camera
	UprightOnly     bool    // Whether to lock the model upright and rotate only around Y
	ShowAxes        bool    // Whether to display axes
	AmbientColor    Color   // Ambient light color
	DiffuseColor    Color   // Diffuse light color
	SpecularColor   Color   // Specular light color
	Shininess       float64 // Shininess factor for specular reflection
	EnableLogging   bool    // Whether to enable logging
	BackfaceCulling bool    // Whether to skip faces pointed away from the camera
}

// Option represents a configuration option.
type Option interface {
	set(*Options)
}

type optionFunc func(*Options)

func (f optionFunc) set(o *Options) {
	f(o)
}

func defaultOptions() *Options {
	return &Options{
		RotationStep:    0.1,
		ZoomScale:       20.0,
		UprightOnly:     false,
		ShowAxes:        true,
		AmbientColor:    Color{R: 0.5, G: 0.5, B: 0.5},
		DiffuseColor:    Color{R: 0.7, G: 0.7, B: 0.7},
		SpecularColor:   Color{R: 1.0, G: 1.0, B: 1.0},
		Shininess:       32.0,
		EnableLogging:   false,
		BackfaceCulling: true,
	}
}

// RotationStep sets the rotation step size.
func RotationStep(step float64) Option {
	return optionFunc(func(o *Options) {
		o.RotationStep = step
	})
}

// ZoomScale sets the initial zoom scale used when projecting the model.
func ZoomScale(scale float64) Option {
	return optionFunc(func(o *Options) {
		o.ZoomScale = scale
	})
}

// UprightOnly locks the model upright so it cannot pitch or roll upside down.
func UprightOnly(enable bool) Option {
	return optionFunc(func(o *Options) {
		o.UprightOnly = enable
	})
}

// ShowAxes sets whether to display axes.
func ShowAxes(show bool) Option {
	return optionFunc(func(o *Options) {
		o.ShowAxes = show
	})
}

// AmbientColor sets the ambient light color.
func AmbientColor(color Color) Option {
	return optionFunc(func(o *Options) {
		o.AmbientColor = color
	})
}

// DiffuseColor sets the diffuse light color.
func DiffuseColor(color Color) Option {
	return optionFunc(func(o *Options) {
		o.DiffuseColor = color
	})
}

// SpecularColor sets the specular light color.
func SpecularColor(color Color) Option {
	return optionFunc(func(o *Options) {
		o.SpecularColor = color
	})
}

// Shininess sets the shininess factor for specular reflection.
func Shininess(shininess float64) Option {
	return optionFunc(func(o *Options) {
		o.Shininess = shininess
	})
}

// EnableLogging sets whether to enable logging.
func EnableLogging(enable bool) Option {
	return optionFunc(func(o *Options) {
		o.EnableLogging = enable
	})
}

// BackfaceCulling sets whether faces pointed away from the camera are skipped.
func BackfaceCulling(enable bool) Option {
	return optionFunc(func(o *Options) {
		o.BackfaceCulling = enable
	})
}
