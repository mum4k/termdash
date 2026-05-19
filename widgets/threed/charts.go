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

package threed

// GenerateLineChartModel generates a 3D line chart model from data.
func GenerateLineChartModel(data []float64) *Model {
	model := NewModel()
	numPoints := len(data)
	vertices := make([]Vector3D, numPoints)
	for i, value := range data {
		x := float64(i) - float64(numPoints)/2
		vertices[i] = Vector3D{X: x, Y: value, Z: 0}
	}
	for i := 0; i < numPoints-1; i++ {
		model.AddFace(Face{
			Vertices: []Vector3D{vertices[i], vertices[i+1]},
			Char:     '─',
		})
	}
	return model
}

// GenerateBarChartModel generates a 3D bar chart model from data.
func GenerateBarChartModel(data []float64) *Model {
	model := NewModel()
	numBars := len(data)
	for i, value := range data {
		x := float64(i) - float64(numBars)/2
		bar := CreateCube(Vector3D{X: x, Y: 0, Z: 0}, value, '█')
		model.Append(bar)
	}
	return model
}
