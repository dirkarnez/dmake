package rpp

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strings"
)

type IElement interface {
	writeString(ownProperty string) string
	getOwnProperty() string
}

type element struct {
	tagName  string
	children []IElement
}

var numOfTabs = 0

func getTabs() string {
	b := make([]byte, numOfTabs)
	for i := range b {
		b[i] = '\t'
	}
	return string(b)
}

func (e *element) writeString(ownProperty string) string {
	var str = getTabs() + "<" + e.tagName + "\n"
	numOfTabs++
	str = str + getTabs() + ownProperty + "\n"
	for _, child := range e.children {
		str = str + child.writeString(child.getOwnProperty())
	}
	numOfTabs--
	str = str + getTabs() + ">" + "\n"
	return str
}

type project struct {
	element

	//AutoXFade string `rpp:"tag:AUTOXFADE;default:0"`
	AutoXFade bool `rpp:"AUTOXFADE"`
	// Ripple    bool   `rpp:"RIPPLE"`

	// GROUPOVERRIDE  0 0 0 GROUPOVERRIDE
	// ENVATTACH 1 ENVATTACH
	// POOLEDENVATTACH 0 POOLEDENVATTACH
	// MIXERUIFLAGS 11 48 MIXERUIFLAGS
	// PEAKGAIN 1  PEAKGAIN
	// FEEDBACK 0 FEEDBACK
	// PANLAW 1 PANLAW
	// PROJOFFS 0 0 0 PROJOFFS
	// MAXPROJLEN 0 600 MAXPROJLEN
	// GRID 3199 8 1 8 1 0 0 0 GRID
	// TIMEMODE 1 5 -1 30 0 0 -1 TIMEMODE
	// VIDEO_CONFIG 0 0 256 VIDEO_CONFIG
	// PANMODE 3 PANMODE
	// CURSOR 0 CURSOR
	// ZOOM 100 0 0 ZOOM
	// VZOOMEX 6 0 VZOOMEX
	// USE_REC_CFG 0 USE_REC_CFG
	// RECMODE 1 RECMODE
	// SMPTESYNC 0 30 100 40 1000 300 0 0 1 0 0 SMPTESYNC
	// LOOP 0 LOOP
	// LOOPGRAN 0 4 LOOPGRAN
	// RECORD_PATH "" "" RECORD_PATH

}

func NewProject() *project {
	p := project{}
	p.tagName = "REAPER_PROJECT"
	p.children = []IElement{}
	return &p
}

func (p *project) getOwnProperty() string {
	return __getOwnProperty(p)
}

func (p *project) AddTrack(track *track) {
	fmt.Println(track.getOwnProperty())
	p.children = append(p.children, track)
}

func (p *project) WriteFile(fileName string) error {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	newWriter := bufio.NewWriter(file)
	fmt.Fprint(newWriter, p.writeString(p.getOwnProperty()))
	newWriter.Flush()
	return nil
}

type track struct {
	element
	FreeMode    bool `rpp:"FREEMODE"`
	Volume      float64
	Pan         float64
	InvertPhase bool
}

func NewTrack() *track {
	t := track{}
	t.tagName = "TRACK"
	return &t
}

func (t *track) getOwnProperty() string {
	return __getOwnProperty(t)
}

func __getOwnProperty(obj interface{}) string {
	reg := []string{}

	s := reflect.ValueOf(obj).Elem()
	typeOfT := s.Type()

	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		//fmt.Printf("%d: %s %s\n", i, typeOfT.Field(i).Name, f.Type())
		tagName, ok := typeOfT.Field(i).Tag.Lookup("rpp")
		if ok {
			itf := f.Interface()
			switch v := itf.(type) {
			case bool:
				if v {
					reg = append(reg, fmt.Sprintf("%s 1", tagName))
				} else {
					reg = append(reg, fmt.Sprintf("%s 0", tagName))
				}
			default:
				reg = append(reg, fmt.Sprintf("%s %v", tagName, itf))
			}
		}
	}

	return strings.Join(reg[:], ",")
}
