package transformer_test

import (
	"bytes"
	"centauri-carbon-proxy/transformer"
	"testing"
)

func TestTransformer_Transform(t *testing.T) {
	tests := []struct {
		name        string // description of this test case
		src         []byte
		flush       bool
		transformer transformer.Transformer
		want        []byte
	}{
		{
			name:        "should not return the last n-1 bytes if flush is false",
			src:         []byte("Badger, Snake, Mushroom"),
			flush:       false,
			transformer: transformer.Transformer{Find: []byte("Snake"), Replace: []byte("Badger")},
			want:        []byte("Badger, Badger, Mush"),
		},
		{
			name:        "should return all bytes if flush is true",
			src:         []byte("Badger, Snake, Mushroom"),
			flush:       true,
			transformer: transformer.Transformer{Find: []byte("Snake"), Replace: []byte("Badger")},
			want:        []byte("Badger, Badger, Mushroom"),
		},
		{
			name:        "should return all bytes if find is empty",
			src:         []byte("Badger, Snake, Mushroom"),
			flush:       true,
			transformer: transformer.Transformer{Find: []byte{}, Replace: []byte("Badger")},
			want:        []byte("Badger, Snake, Mushroom"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.transformer.Transform(test.src, test.flush)
			if !bytes.Equal(got, test.want) {
				t.Errorf("Transform() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestTransformer_Transform_Multirun(t *testing.T) {
	src := []byte("Badger, Snake, Mushroom")
	transformer := transformer.Transformer{Find: []byte("Snake"), Replace: []byte("Badger")}

	result1 := transformer.Transform(src, false)
	result2 := transformer.Transform(src, false)
	finalResult := append(result1, result2...)
	result3 := transformer.Transform(src, true)
	finalResult = append(finalResult, result3...)

	want := []byte("Badger, Badger, MushroomBadger, Badger, MushroomBadger, Badger, Mushroom")
	if !bytes.Equal(finalResult, want) {
		t.Errorf("Process() = %v, want %v", finalResult, want)
	}

}

func TestChain_Process(t *testing.T) {
	tests := []struct {
		name  string // description of this test case
		src   []byte
		flush bool
		chain transformer.Chain
		want  []byte
	}{
		{
			name:  "Should run multiple transformers",
			src:   []byte("The quick brown fox jumps over the lazy dog"),
			flush: true,
			chain: transformer.Chain{Transformers: []transformer.Transformer{
				{
					Find:    []byte("quick"),
					Replace: []byte("slow"),
				},
				{
					Find:    []byte("brown"),
					Replace: []byte("blue"),
				},
				{
					Find:    []byte("fox"),
					Replace: []byte("honey badger"),
				},
			}},
			want: []byte("The slow blue honey badger jumps over the lazy dog"),
		},
		{
			name:  "Should run transformers in order",
			src:   []byte("hello"),
			flush: true,
			chain: transformer.Chain{Transformers: []transformer.Transformer{
				{
					Find:    []byte("hello"),
					Replace: []byte("aloha"),
				},
				{
					Find:    []byte("aloha"),
					Replace: []byte("hola"),
				},
				{
					Find:    []byte("hola"),
					Replace: []byte("hi"),
				},
			}},
			want: []byte("hi"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.chain.Process(test.src, test.flush)
			if !bytes.Equal(got, test.want) {
				t.Errorf("Process() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestChain_Process_Multirun(t *testing.T) {
	src1 := []byte("Badger, Snake, Mushroom")
	src2 := []byte("Snake, Chicken, Egg")
	src3 := []byte("Chicken, Egg, Badger")
	src4 := []byte("Chicken, Snake, Mushroom")
	transformers := []transformer.Transformer{
		{Find: []byte("Snake"), Replace: []byte("Hiss")},
		{Find: []byte("Badger"), Replace: []byte("Chomp")},
		{Find: []byte("Egg"), Replace: []byte("Potato")},
	}
	chain := transformer.Chain{
		Transformers: transformers,
	}

	result1 := chain.Process(src1, false)
	result2 := chain.Process(src2, false)
	finalResult := append(result1, result2...)
	result3 := chain.Process(src3, false)
	finalResult = append(finalResult, result3...)
	result4 := chain.Process(src4, true)
	finalResult = append(finalResult, result4...)

	want := []byte("Chomp, Hiss, MushroomHiss, Chicken, PotatoChicken, Potato, ChompChicken, Hiss, Mushroom")
	if !bytes.Equal(finalResult, want) {
		t.Errorf("Process() = %v, want %v", finalResult, want)
	}

}
