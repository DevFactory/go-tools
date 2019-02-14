/* Copyright 2019 DevFactory FZ LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License. */

package collections

// GetSlicesDifferences takes two slices of the same type and returns a slice of elements, that
// are present only in the first slice and another slice of elements that are present only in
// the second slice
func GetSlicesDifferences(first, second []interface{}, equalityFunc func(first, second interface{}) bool) (
	inFirstOnly, inSecondOnly []interface{}) {
	inFirstOnly = []interface{}{}
	inSecondOnly = second
	for _, fromFirst := range first {
		found := -1
		for i, fromSecond := range second {
			if equalityFunc(fromFirst, fromSecond) {
				found = i
				break
			}
		}
		if found > -1 {
			// remove from inSecondOnly
			inSecondOnly = append(inSecondOnly[:found], inSecondOnly[found+1:]...)
		} else {
			inFirstOnly = append(inFirstOnly, fromFirst)
		}
	}
	return inFirstOnly, inSecondOnly
}
