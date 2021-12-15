// package main

// import (
// 	"bufio"
// 	"fmt"
// 	"log"
// 	"os"

// 	"github.com/dirkarnez/rpp/rpp"
// )

// func main() {
// 	f, err := os.Open("test.rpp")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	reader := bufio.NewReader(f)
// 	project, err := rpp.ParseRPP(reader)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	for _, track := range project.Tracks {
// 		fmt.Printf("%s\n", track.Name)
// 		if track.Name == "Print" {
// 			continue
// 		}

// 		fmt.Printf("  Volume = %f\n", track.Volume)
// 		fmt.Printf("  Pan = %f\n", track.Pan)

// 		if track.FXChain != nil {
// 			fmt.Printf("  FX\n")
// 			for _, fx := range track.FXChain.FX {
// 				if fx.VST != nil {
// 					data := fx.VST.Data
// 					fmt.Printf("    %s\n", fx.VST.Path)
// 					if fx.VST.ReaEQ != nil {
// 						for _, band := range fx.VST.ReaEQ.Bands {
// 							//fmt.Printf("      [%d] freq=%7.1f Hz, gain=%6.2f dB, bw=%5.3f, q=%6.3f\n", i, band.Frequency, band.Gain, band.Bandwidth, band.Q())
// 							fmt.Printf("      {\"freq\":%f, \"gain\":%f, \"q\":%f},\n", band.Frequency, band.Gain, band.Q())
// 						}
// 					} else {
// 						fmt.Printf("      %2X\n", data)
// 					}
// 				} else if fx.JS != nil {
// 					fmt.Printf("    %s\n", fx.JS.Path)
// 				}
// 			}
// 		}
// 	}

// }

package main

import (
	"encoding/binary"
	"log"
	"os"

	"github.com/alecthomas/kong"
	"github.com/alecthomas/repr"
	"github.com/dirkarnez/dmake/generator/rpp"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type File struct {
	Elements []*Element `@@*`
}

type Element struct {
	ElementHavingChildren   *ElementHavingChildren   `  @@`
	ElementHavingNoChildren *ElementHavingNoChildren `| @@`
}

type ElementHavingNoChildren struct {
	Name       string       `"<" @Ident`
	Attributes []*Attribute `@@* "/" ">"`
}

type ElementHavingChildren struct {
	Name       string       `"<" @Ident`
	Attributes []*Attribute `@@* ">"`
	Children   []*Element   `@@*`
	NameEnd    string       `"<" "/" @Ident ">"`
}

type Attribute struct {
	Key   string `@Ident"="`
	Value *Value `@@`
}

type Value struct {
	String *string  `  @String`
	Number *float64 `| @Float`
}

var (
	graphQLLexer = lexer.Must(lexer.NewSimple([]lexer.Rule{
		{Name: "Ident", Pattern: `[a-zA-Z]+`, Action: nil},
		{Name: "String", Pattern: `"(?:\\.|[^"])*"`, Action: nil},
		{Name: "Float", Pattern: `[-+]?\d+(?:\.\d+)?`, Action: nil},
		{Name: "Punct", Pattern: `[-,()*/+%{};&!=:<>]|\[|\]`, Action: nil},
		{Name: "Whitespace", Pattern: `[ \t\n\r]+`, Action: nil},
	}))
	parser = participle.MustBuild(&File{},
		participle.Lexer(graphQLLexer),
		participle.Elide("Whitespace"),
		participle.Unquote("String"),
		participle.UseLookahead(50),
	)
)

var cli struct {
	EBNF  bool     `help"Dump EBNF."`
	Files []string `arg:"" optional:"" type:"existingfile" help:"GraphQL schema files to parse."`
}

func main() {
	ctx := kong.Parse(&cli)
	// if cli.EBNF {
	// 	fmt.Println(parser.String())
	// 	ctx.Exit(0)
	// }
	ast := &File{}
	r, err := os.Open("samples/band.dmake")
	ctx.FatalIfErrorf(err)
	defer r.Close()
	err = parser.Parse("", r, ast)
	ctx.FatalIfErrorf(err)
	repr.Println(ast)

	track1 := rpp.NewTrack()
	track1.FreeMode = false

	track2 := rpp.NewTrack()
	track2.FreeMode = true

	project := rpp.NewProject()
	project.AutoXFade = true
	project.AddTrack(track1)
	project.AddTrack(track2)

	project.WriteFile("generated.rpp")

	dump()
}

const (
	Byte                   int32 = 0
	ByteEnabled                int32 = 0
	ByteNoteOn                 int32 = 1
	ByteVol                     int32 = 2
	BytePan                     int32 = 3
	ByteMidiChan                int32 = 4
	ByteMidiNote                int32 = 5
	ByteMidiPatch               int32 = 6
	ByteMidiBank                int32 = 7
	ByteLoopActive              int32 = 9
	ByteShowInfo                int32 = 10
	ByteShuffle                 int32 = 11
	ByteMainVol                 int32 = 12
	ByteStretch                 int32 = 13
	BytePitchable               int32 = 14
	ByteZipped                  int32 = 15
	ByteDelayFlags              int32 = 16
	BytePatLength               int32 = 17
	ByteBlockLength             int32 = 18
	ByteUseLoopPoints           int32 = 19
	ByteLoopType                int32 = 20
	ByteChanType                int32 = 21
	ByteMixSliceNum             int32 = 22
	ByteEffectChannelMuted      int32 = 27
	BytePlayTruncatedNotes      int32 = 30

	Word              int32 = 64
	WordNewChan       int32 = Word
	WordNewPat        int32 = Word + 1
	WordTempo         int32 = Word + 2
	WordCurrentPatNum int32 = Word + 3
	WordPatData       int32 = Word + 4
	WordFx            int32 = Word + 5
	WordFadeStereo    int32 = Word + 6
	WordCutOff        int32 = Word + 7
	WordDotVol        int32 = Word + 8
	WordDotPan        int32 = Word + 9
	WordPreAmp        int32 = Word + 10
	WordDecay         int32 = Word + 11
	WordAttack        int32 = Word + 12
	WordDotNote       int32 = Word + 13
	WordDotPitch      int32 = Word + 14
	WordDotMix        int32 = Word + 15
	WordMainPitch     int32 = Word + 16
	WordRandChan      int32 = Word + 17
	WordMixChan       int32 = Word + 18
	WordResonance     int32 = Word + 19
	WordLoopBar       int32 = Word + 20
	WordStDel         int32 = Word + 21
	WordFx3           int32 = Word + 22
	WordDotReso       int32 = Word + 23
	WordDotCutOff     int32 = Word + 24
	WordShiftDelay    int32 = Word + 25
	WordLoopEndBar    int32 = Word + 26
	WordDot           int32 = Word + 27
	WordDotShift      int32 = Word + 28
	WordLayerChans    int32 = Word + 30
	WordInsertIcon    int32 = Word + 31
	WordCurrentSlotNum int32 = Word + 34

	Int                int32 = 128
	DWordColor         int32 = Int
	DWordPlayListItem  int32 = Int + 1
	DWordEcho          int32 = Int + 2
	DWordFxSine        int32 = Int + 3
	DWordCutCutBy      int32 = Int + 4
	DWordWindowH       int32 = Int + 5
	DWordMiddleNote    int32 = Int + 7
	DWordReserved      int32 = Int + 8
	DWordMainResoCutOff int32 = Int + 9
	DWordDelayReso     int32 = Int + 10
	DWordReverb        int32 = Int + 11
	DWordIntStretch    int32 = Int + 12
	DWordSsNote        int32 = Int + 13
	DWordFineTune      int32 = Int + 14
	DWordInsertColor   int32 = Int + 21
	DWordFineTempo     int32 = Int + 28

	Undef             int32 = 192
	Text              int32 = Undef
	TextChanName      int32 = Text
	TextPatName       int32 = Text + 1
	TextTitle         int32 = Text + 2
	TextComment       int32 = Text + 3
	TextSampleFileName int32 = Text + 4
	TextUrl           int32 = Text + 5
	TextCommentRtf    int32 = Text + 6
	TextVersion       int32 = Text + 7
	GeneratorName     int32 = Text + 9
	TextPluginName    int32 = Text + 11
	TextInsertName    int32 = Text + 12
	TextGenre         int32 = Text + 14
	TextAuthor        int32 = Text + 15
	TextMidiCtrls     int32 = Text + 16
	TextDelay         int32 = Text + 17

	Data                  int32 = 210
	DataTs404Params       int32 = Data
	DataDelayLine         int32 = Data + 1
	DataNewPlugin         int32 = Data + 2
	DataPluginParams      int32 = Data + 3
	DataChanParams        int32 = Data + 5
	DataEnvLfoParams      int32 = Data + 8
	DataBasicChanParams   int32 = Data + 9
	DataOldFilterParams   int32 = Data + 10
	DataOldAutomationData int32 = Data + 13
	DataPatternNotes      int32 = Data + 14
	DataInsertParams      int32 = Data + 15
	DataAutomationChannels int32 = Data + 17
	DataChanGroupName     int32 = Data + 21
	DataPlayListItems     int32 = Data + 23
	DataAutomationData    int32 = Data + 24
	DataInsertRoutes      int32 = Data + 25
	DataInsertFlags       int32 = Data + 26
	DataSaveTimestamp     int32 = Data + 27
)

type Chunk struct {
	Event int32
}

type FL struct {
	A [4]byte
	B int32
	C int16
	Channel_count int16
	Ppq int16
	D [4]byte
	ChunkLength int32

}

func dump() {
    f, err := os.Create("file.bin")
    if err != nil {
        log.Fatal("Couldn't open file")
    }
    defer f.Close()
	var FLhd [4]byte
	copy(FLhd[:], "FLhd")

	var FLdt [4]byte
	copy(FLdt[:], "FLdt")
	var chunkLength int32 = 113566
	var chunks = [1]Chunk{ {Event: TextVersion} }
	log.Println(chunks[0].Event)
    var data = FL{FLhd, 6, 0, 1, 96, FLdt, chunkLength}
    err = binary.Write(f, binary.LittleEndian, data)
    if err != nil {
        log.Fatal("Write failed")
    }
}