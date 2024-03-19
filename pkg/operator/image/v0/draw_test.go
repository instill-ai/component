package image

import (
	"encoding/json"
	"testing"

	_ "embed" // embed

	"google.golang.org/protobuf/types/known/structpb"
)

var (
	//go:embed testdata/cls-dog.json
	clsDogJSON []byte
	//go:embed testdata/det-coco-1.json
	detCOCO1JSON []byte
	//go:embed testdata/det-coco-2.json
	detCOCO2JSON []byte
	//go:embed testdata/kp-coco-1.json
	kpCOCO1JSON []byte
	//go:embed testdata/kp-coco-2.json
	kpCOCO2JSON []byte
	//go:embed testdata/ocr-mm.json
	ocrMMJSON []byte
	//go:embed testdata/inst-seg-coco-1.json
	instSegCOCO1JSON []byte
	//go:embed testdata/inst-seg-coco-2.json
	instSegCOCO2JSON []byte
	//go:embed testdata/inst-seg-stomata.json
	instSegStomataJSON []byte
	//go:embed testdata/sem-seg-cityscape.json
	semSegCityscapeJSON []byte
)

// TestDrawClassification tests the drawClassification function
func TestDrawClassification(t *testing.T) {

	inputDog := &structpb.Struct{}
	if err := json.Unmarshal(clsDogJSON, inputDog); err != nil {
		panic(err)
	}

	inputs := []*structpb.Struct{
		inputDog,
	}

	e := &Execution{}
	e.Task = "TASK_DRAW_CLASSIFICATION"

	if _, err := e.Execute(inputs); err != nil {
		t.Fatalf("drawClassification returned an error: %v", err)
	}
}

// TestDrawDetection tests the drawDetection function
func TestDrawDetection(t *testing.T) {

	inputCOCO1 := &structpb.Struct{}
	if err := json.Unmarshal(detCOCO1JSON, inputCOCO1); err != nil {
		panic(err)
	}

	inputCOCO2 := &structpb.Struct{}
	if err := json.Unmarshal(detCOCO2JSON, inputCOCO2); err != nil {
		panic(err)
	}

	inputs := []*structpb.Struct{
		inputCOCO1,
		// inputCOCO2,
	}

	e := &Execution{}
	e.Task = "TASK_DRAW_DETECTION"

	if _, err := e.Execute(inputs); err != nil {
		t.Fatalf("drawDetection returned an error: %v", err)
	}
}

// TestDrawKeypoint tests the drawKeypoint function
func TestDrawKeypoint(t *testing.T) {

	inputCOCO1 := &structpb.Struct{}
	if err := json.Unmarshal(kpCOCO1JSON, inputCOCO1); err != nil {
		panic(err)
	}

	inputCOCO2 := &structpb.Struct{}
	if err := json.Unmarshal(kpCOCO2JSON, inputCOCO2); err != nil {
		panic(err)
	}

	inputs := []*structpb.Struct{
		inputCOCO1,
		// inputCOCO2,
	}

	e := &Execution{}
	e.Task = "TASK_DRAW_KEYPOINT"

	if _, err := e.Execute(inputs); err != nil {
		t.Fatalf("drawKeypoint returned an error: %v", err)
	}
}

// TestDrawOCR tests the drawOCR function
func TestDrawOCR(t *testing.T) {

	inputMM := &structpb.Struct{}
	if err := json.Unmarshal(ocrMMJSON, inputMM); err != nil {
		panic(err)
	}

	inputCOCO2 := &structpb.Struct{}
	if err := json.Unmarshal(kpCOCO2JSON, inputCOCO2); err != nil {
		panic(err)
	}

	inputs := []*structpb.Struct{
		inputMM,
	}

	e := &Execution{}
	e.Task = "TASK_DRAW_OCR"

	if _, err := e.Execute(inputs); err != nil {
		t.Fatalf("drawKeypoint returned an error: %v", err)
	}
}

// TestDrawInstanceSegmentation tests the drawInstanceSegmentation function
func TestDrawInstanceSegmentation(t *testing.T) {

	inputCOCO1 := &structpb.Struct{}
	if err := json.Unmarshal(instSegCOCO1JSON, inputCOCO1); err != nil {
		panic(err)
	}

	inputCOCO2 := &structpb.Struct{}
	if err := json.Unmarshal(instSegCOCO2JSON, inputCOCO2); err != nil {
		panic(err)
	}

	inputStomata := &structpb.Struct{}
	if err := json.Unmarshal(instSegStomataJSON, inputStomata); err != nil {
		panic(err)
	}

	inputs := []*structpb.Struct{
		inputCOCO1,
		inputCOCO2,
		inputStomata,
	}

	e := &Execution{}
	e.Task = "TASK_DRAW_INSTANCE_SEGMENTATION"

	if _, err := e.Execute(inputs); err != nil {
		t.Fatalf("drawInstanceSegmentation returned an error: %v", err)
	}
}

// TestDrawSemanticSegmentation tests the drawSemanticSegmentation function
func TestDrawSemanticSegmentation(t *testing.T) {

	inputCityscape := &structpb.Struct{}
	if err := json.Unmarshal(semSegCityscapeJSON, inputCityscape); err != nil {
		panic(err)
	}

	inputs := []*structpb.Struct{
		inputCityscape,
	}

	e := &Execution{}
	e.Task = "TASK_DRAW_SEMANTIC_SEGMENTATION"

	if _, err := e.Execute(inputs); err != nil {
		t.Fatalf("drawSemanticSegmentation returned an error: %v", err)
	}
}
