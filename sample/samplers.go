package sample

import (
	"errors"
	"math"

	"golang.org/x/exp/rand"
	"gonum.org/v1/gonum/stat/sampleuv"
)

type Sampler interface {
	Sample([]float32) (int32, error)
}

type tokenInfo struct {
	id    int
	logit float64
	prob  float64
}

type tokenSliceInfo struct {
	tokens []tokenInfo
	sorted bool
}

type weighted struct {
	src        rand.Source
	transforms []Transform
}

// TODO(parthsareen): remove uv sample dependency https://github.com/ollama/ollama/issues/9279
func Weighted(seed *uint64, transforms ...Transform) Sampler {
	var src rand.Source
	if seed != nil {
		src = rand.NewSource(*seed)
	}
	return weighted{src: src, transforms: transforms}
}

func (s weighted) Sample(logits []float32) (int32, error) {
	logits64 := make([]float64, len(logits))
	for i, v := range logits {
		logits64[i] = float64(v)
	}

	probs := softmax(logits64)

	tokens := make([]tokenInfo, len(logits))
	for i, v := range logits {
		tokens[i] = tokenInfo{
			id:    i,
			logit: float64(v),
			prob:  probs[i],
		}
	}

	tokensInfo := tokenSliceInfo{tokens: tokens, sorted: false}
	for _, t := range s.transforms {
		tokensInfo = t.Apply(tokensInfo)
	}

	if len(tokensInfo.tokens) == 0 {
		return -1, errors.New("no valid logits found for weighed sampling")
	}

	filteredProbs := make([]float64, len(tokensInfo.tokens))
	indices := make([]int, len(tokensInfo.tokens))
	for i, token := range tokensInfo.tokens {
		filteredProbs[i] = token.prob
		indices[i] = token.id
	}

	w := sampleuv.NewWeighted(filteredProbs, s.src)
	if idx, ok := w.Take(); ok {
		return int32(indices[idx]), nil
	}
	return -1, errors.New("weighed sampler failed, no valid token found")
}

type greedy struct {
	transforms []Transform
}

func Greedy() Sampler {
	return greedy{}
}

func (s greedy) Sample(logits []float32) (int32, error) {
	logits64 := make([]float64, len(logits))
	for i, v := range logits {
		logits64[i] = float64(v)
	}

	var maxIdx int
	var maxLogit float64
	for i, logit := range logits64 {
		if logit > maxLogit {
			maxLogit = logit
			maxIdx = i
		}
	}

	if maxLogit == math.Inf(-1) {
		return -1, errors.New("no valid logits found for greedy sampling")
	}

	return int32(maxIdx), nil
}

// TODO(parthsareen): update sampler interface to use json unmarshal https://github.com/ollama/ollama/issues/9278
func NewSampler(temperature float32, topK int, topP float32, minP float32, seed int) (Sampler, error) {
	transforms := []Transform{}
	if temperature < 0 || temperature > 2 {
		return nil, errors.New("temperature must be between 0 and 2")
	}

	if temperature != 0 {
		transforms = append(transforms, Temperature(temperature))
	}

	if topK != 0 {
		if topK <= 0 {
			return nil, errors.New("topK must be greater than 0")
		}
		transforms = append(transforms, TopK(topK))
	}

	if topP != 0 {
		if topP < 0 || topP >= 1 {
			return nil, errors.New("topP must be between 0 and 1")
		}
		transforms = append(transforms, TopP(topP))
	}

	if minP != 0 {
		if minP < 0 || minP >= 1 {
			return nil, errors.New("minP must be between 0 and 1")
		}
		transforms = append(transforms, MinP(minP))
	}

	if len(transforms) == 0 {
		return nil, errors.New("at least one transform is required")
	}

	if temperature == 0 {
		return Greedy(), nil
	}

	if seed != 0 {
		seed64 := uint64(seed)
		return Weighted(&seed64, transforms...), nil
	}
	return Weighted(nil, transforms...), nil
}
