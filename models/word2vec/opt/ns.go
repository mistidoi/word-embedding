// Copyright © 2017 Makoto Ito
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package opt

import (
	"github.com/ynqa/word-embedding/models"
	"github.com/ynqa/word-embedding/models/word2vec"
	"github.com/ynqa/word-embedding/utils"
	"github.com/ynqa/word-embedding/utils/fileio"
	"github.com/ynqa/word-embedding/utils/vector"
)

type NegativeSampling struct {
	models.Common
	SampleSize int
}

func (ns NegativeSampling) PreTrain() error {
	learningRate = ns.LearningRate
	word2vec.GlobalFreqMap = utils.NewFreqMap()

	if err := fileio.Load(ns.Common.InputFile, word2vec.GlobalFreqMap.Update); err != nil {
		return err
	}

	word2vec.GlobalWordMap = word2vec.NewWordMapFrom(word2vec.GlobalFreqMap.Keys(), ns.Common.Dimension, true)
	return nil
}

func (ns NegativeSampling) Update(target string, contentVectors, poolVector vector.Vector) {
	var label int
	var negativeVector vector.Vector
	var randTerm string
	var randWord *word2vec.Word
	for n := -1; n < ns.SampleSize; n++ {
		if n == -1 {
			label = 1
			negativeVector = word2vec.GlobalWordMap[target].VectorAsNegative
		} else {
			label = 0
			randTerm, randWord = word2vec.GlobalWordMap.GetRandom()
			if randTerm == target {
				continue
			}
			negativeVector = randWord.VectorAsNegative
		}

		f := utils.Sigmoid(negativeVector.Inner(contentVectors))
		g := (float64(label) - f) * learningRate

		for d := 0; d < ns.Common.Dimension; d++ {
			poolVector[d] += g * negativeVector[d]
			negativeVector[d] += g * contentVectors[d]
		}

		if n == -1 {
			word2vec.GlobalWordMap[target].VectorAsNegative = negativeVector
		} else {
			word2vec.GlobalWordMap[randTerm].VectorAsNegative = negativeVector
		}
	}
}
