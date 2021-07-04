package main

// #cgo LDFLAGS: -L${SRCDIR}/../govarnam -lgovarnam
// #cgo CFLAGS: -I${SRCDIR}/../govarnam -DHAVE_SNPRINTF -DPREFER_PORTABLE_SNPRINTF -DNEED_ASPRINTF
// #include <libgovarnam.h>
import "C"

import (
	"context"
	"log"
	"unsafe"
)

var gCtx context.Context

// Suggestion suggestion
type Suggestion struct {
	Word      string
	Weight    int
	LearnedOn int
}

// TransliterationResult result
type TransliterationResult struct {
	ExactMatch            []Suggestion
	Suggestions           []Suggestion
	GreedyTokenized       []Suggestion
	DictionaryResultCount int
}

// Convert a C Suggestion to Go
func makeSuggestion(cSug C.struct_Suggestion_t) Suggestion {
	var sug Suggestion
	sug.Word = C.GoString(cSug.Word)
	sug.Weight = int(cSug.Weight)
	sug.LearnedOn = int(cSug.LearnedOn)
	return sug
}

func makeGoTransliterationResult(ctx *context.Context, cResults *C.struct_TransliterationResult_t) TransliterationResult {
	var results TransliterationResult

	select {
	case <-(*ctx).Done():
		return results
	default:
		var i int

		var exactMatch []Suggestion
		i = 0
		for i < int(C.varray_length(cResults.ExactMatch)) {
			cSug := *(*C.Suggestion)(C.varray_get(cResults.ExactMatch, C.int(i)))
			sug := makeSuggestion(cSug)
			exactMatch = append(exactMatch, sug)
			i++
		}
		results.ExactMatch = exactMatch

		var suggestions []Suggestion
		i = 0
		for i < int(C.varray_length(cResults.Suggestions)) {
			cSug := *(*C.Suggestion)(C.varray_get(cResults.Suggestions, C.int(i)))
			sug := makeSuggestion(cSug)
			suggestions = append(suggestions, sug)
			i++
		}
		results.Suggestions = suggestions

		var greedyTokenized []Suggestion
		i = 0
		for i < int(C.varray_length(cResults.GreedyTokenized)) {
			cSug := *(*C.Suggestion)(C.varray_get(cResults.Suggestions, C.int(i)))
			sug := makeSuggestion(cSug)
			greedyTokenized = append(greedyTokenized, sug)
			i++
		}
		results.GreedyTokenized = greedyTokenized
		return results
	}
}

func initVarnam(id string) {
	err := C.varnam_init_from_id(C.CString(id))
	if err != C.VARNAM_SUCCESS {
		logVarnamError()
	}
}

func debugVarnam(val bool) {
	if *debug {
		C.varnam_debug(C.int(1))
	}
}

func setConfigVarnam(config *Config) {
	C.varnam_set_dictionary_suggestions_limit(C.int(config.DictionarySuggestionsLimit))
	C.varnam_set_tokenizer_suggestions_limit(C.int(config.TokenizerSuggestionsLimit))
}

func transliterateWithContext(ctx *context.Context, word string) TransliterationResult {
	p := unsafe.Pointer(ctx)

	cResult := C.TransliterateWithContext(p, C.CString(word))

	return makeGoTransliterationResult(ctx, cResult)
}

func learnWordVarnam(word string) {
	C.varnam_learn(C.CString(word), 0)
}

func logVarnamError() {
	log.Fatal(C.GoString(C.varnam_get_last_error()))
}
